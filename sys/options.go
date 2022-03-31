package sys

import (
	"strconv"
	"strings"
	"time"

	"github.com/ppkg/microgo/utils"

	gwruntime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

	"runtime"

	"github.com/maybgit/glog"
)

var (
	_opt         *Options
	GlobalConfig SystemGlobalConfig //consul kv
)

type Options struct {
	Name          string //service name
	ConsulAddress string //consul server address
	LocalIP       string //local ip address
	HttpPort      *int   //http port default 0 dynamic
	GrpcPort      *int   //grpc port default 0 dynamic
	PprofPort     *int   //pprof port

	XxljobAddress string //xxljob manager address
	XxljobPort    *int   //xxljob execute port

	Tags []string
	Mux  []gwruntime.ServeMuxOption
}

func Init(o *Options) {
	o.ConsulAddress = *consulAddress
	o.GrpcPort = grpcPort
	o.HttpPort = httpPort
	o.PprofPort = pprofPort
	o.XxljobPort = xxljobPort
	o.XxljobAddress = *xxljobAddress

	if o.ConsulAddress == "" {
		if runtime.GOOS == "windows" || runtime.GOOS == "darwin" {
			o.ConsulAddress = "127.0.0.1:8500"
		} else {
			panic("consul address is empty")
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
}

func (o *Options) Init() {
	Init(o)
}

func GetOption(opts ...Options) *Options {
	if len(opts) == 1 {
		opts[0].Init()
	}
	for _opt == nil {
		glog.Info("wait options init...")
		time.Sleep(time.Second)
	}
	return _opt
}
