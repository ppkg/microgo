package sys

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	gwruntime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	consulapi "github.com/hashicorp/consul/api"
	"github.com/ppkg/microgo/utils"

	"github.com/maybgit/glog"
)

var (
	_opt          *Options
	_globalConfig SystemGlobalConfig //consul kv
)

type SystemGlobalConfig struct {
	HdRelease bool //是否灰度
	Debug     bool //调试模式
}

type Options struct {
	Name          string //service name
	ConsulAddress string //consul server address
	LocalIP       string //local ip address
	HttpPort      *int   //http port default 0 dynamic
	GrpcPort      *int   //grpc port default 0 dynamic
	PprofPort     *int   //pprof port
	PprofDisable  bool   //disable pprof

	XxljobAddress string //xxljob manager address
	XxljobPort    *int   //xxljob execute port

	JaegerEndpoint string

	Tags []string
	Mux  []gwruntime.ServeMuxOption

	Ctx context.Context
}

func Init(o *Options) {
	if o.GrpcPort == nil || *grpcPort > 0 {
		o.GrpcPort = grpcPort
	}
	if o.HttpPort == nil || *httpPort > 0 {
		o.HttpPort = httpPort
	}
	if o.PprofPort == nil || *pprofPort > 0 {
		o.PprofPort = pprofPort
	}
	if o.XxljobPort == nil || *xxljobPort > 0 {
		o.XxljobPort = xxljobPort
	}
	if *xxljobAddress != "" {
		o.XxljobAddress = *xxljobAddress
	}

	if !IsLinux() {
		o.ConsulAddress = "127.0.0.1:8500"
	} else {
		if *consulAddress != "" {
			o.ConsulAddress = *consulAddress
		} else if o.ConsulAddress == "" {
			panic("consul address empty")
		}
	}

	if o.LocalIP == "" {
		o.LocalIP = utils.GetIp()
		if o.LocalIP == "" {
			glog.Error("get local ip error")
		} else {
			glog.Info("local ip address", o.LocalIP)
		}
	}

	_opt = o

	go func() {
		time.Sleep(time.Second * 3)
		var sb strings.Builder
		sb.WriteString("\nservice name     : " + _opt.Name)
		sb.WriteString("\nlocal ip         : " + _opt.LocalIP)
		sb.WriteString("\nconsul address   : " + _opt.ConsulAddress)
		sb.WriteString("\nhttp port        : " + strconv.Itoa(*_opt.HttpPort))
		sb.WriteString("\ngrpc port        : " + strconv.Itoa(*_opt.GrpcPort))
		sb.WriteString("\npprof port       : " + strconv.Itoa(*_opt.PprofPort))
		sb.WriteString("\nxxljob address   : " + _opt.XxljobAddress)
		sb.WriteString("\nxxljob port      : " + strconv.Itoa(*_opt.XxljobPort))
		sb.WriteString("\n")
		glog.Info(sb.String())
	}()

	initTracerProvider()
	watchSystemGlobalConfig()
}

func (o *Options) Init() {
	Init(o)
}

func GetOption() *Options {
	for _opt == nil {
		glog.Info("wait options init...")
		time.Sleep(time.Second)
	}
	return _opt
}

func watchSystemGlobalConfig() {
	// watch 服务配置
	go func() {
		var lastIndex uint64
		for {
			opt := GetOption()
			if client, err := consulapi.NewClient(&consulapi.Config{Address: opt.ConsulAddress, Scheme: "http"}); err != nil {
				glog.Warning(err)
			} else {
				pair, meta, err := client.KV().Get("SystemGlobalConfig", &consulapi.QueryOptions{WaitIndex: lastIndex})
				if err != nil {
					glog.Error(err)
				} else {
					if pair != nil {
						lastIndex = meta.LastIndex
						glog.Info("SystemGlobalConfig", string(pair.Value))
						json.Unmarshal(pair.Value, &_globalConfig)
						glog.Info("_globalConfig.HdRelease", _globalConfig.HdRelease)
					} else {
						glog.Error("SystemGlobalConfig pair nil")
					}
				}
			}
			time.Sleep(time.Second * 3)
		}
	}()
}
