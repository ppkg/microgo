package sys

import (
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
	Address       string //service address(default auto)
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

	if o.Address == "" {
		o.Address = utils.GetIp()
		if o.Address == "" {
			glog.Error("get local ip error")
		} else {
			glog.Info("local ip address", o.Address)
		}
	}

	_opt = o
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
