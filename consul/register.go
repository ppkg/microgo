package consul

import (
	"context"
	"fmt"
	"net"

	"github.com/ppkg/microgo/sys"

	"time"

	"github.com/hashicorp/consul/api"
	"github.com/maybgit/glog"
	"google.golang.org/grpc/health/grpc_health_v1"
)

type HealthImpl struct{}

func (h *HealthImpl) Check(ctx context.Context, req *grpc_health_v1.HealthCheckRequest) (*grpc_health_v1.HealthCheckResponse, error) {
	// glog.Info("health checking")
	return &grpc_health_v1.HealthCheckResponse{
		Status: grpc_health_v1.HealthCheckResponse_SERVING,
	}, nil
}

func (h *HealthImpl) Watch(req *grpc_health_v1.HealthCheckRequest, w grpc_health_v1.Health_WatchServer) error {
	return nil
}

// 注册consul grpc
func RegisterGrpcService(listen ...net.Listener) {
	opt := sys.GetOption()
	if len(listen) == 1 {
		*opt.GrpcPort = listen[0].Addr().(*net.TCPAddr).Port
	} else {
		for *opt.GrpcPort == 0 {
			time.Sleep(time.Second * 1)
		}
	}

	consulConfig := api.DefaultConfig()
	consulConfig.Address = opt.ConsulAddress
	client, err := api.NewClient(consulConfig)
	if err != nil {
		glog.Error(err)
		return
	}
	agent := client.Agent()
	interval := time.Second
	deregister := time.Second * 300

	reg := &api.AgentServiceRegistration{
		ID:      fmt.Sprintf("%v-%v:%v", opt.Name, opt.Address, *opt.GrpcPort),
		Name:    opt.Name,
		Port:    *opt.GrpcPort,
		Address: opt.Address,
		Check: &api.AgentServiceCheck{
			Interval:                       interval.String(),
			GRPC:                           fmt.Sprintf("%v:%v/%v", opt.Address, *opt.GrpcPort, opt.Name),
			DeregisterCriticalServiceAfter: deregister.String(),
		},
	}
	reg.Tags = append(reg.Tags, "grpc", opt.Address)
	reg.Tags = append(reg.Tags, opt.Tags...)

	if err := agent.ServiceRegister(reg); err != nil {
		glog.Error("Service Register error ", err)
		return
	} else {
		glog.Info("Success Regitser grpc Service", reg.ID)
	}
}

// 注册consul http
func RegisterHttpService(opt *sys.Options, listen ...net.Listener) {
	var is_pprof bool
	for _, v := range opt.Tags {
		if is_pprof = v == "pprof"; is_pprof {
			break
		}
	}

	var port int
	if len(listen) == 1 {
		port = listen[0].Addr().(*net.TCPAddr).Port
		if is_pprof {
			*opt.PprofPort = port
		} else {
			*opt.HttpPort = port
		}
	} else {
		if is_pprof {
			for *opt.PprofPort == 0 {
				glog.Info("pprof port ", *opt.HttpPort)
				time.Sleep(time.Second * 1)
			}
			port = *opt.PprofPort
		} else {
			for *opt.HttpPort == 0 {
				glog.Info("http port ", *opt.HttpPort)
				time.Sleep(time.Second * 1)
			}
			port = *opt.HttpPort
		}
	}

	glog.Info("port", port)

	config := api.DefaultConfig()
	config.Address = opt.ConsulAddress
	client, err := api.NewClient(config)
	if err != nil {
		glog.Error("consul client error : ", err)
	}

	reg := new(api.AgentServiceRegistration)
	reg.ID = fmt.Sprintf("%v-%v:%v", opt.Name, opt.Address, port)
	reg.Name = opt.Name
	reg.Port = port
	reg.Address = opt.Address

	reg.Tags = append(reg.Tags, "http", opt.Address)
	reg.Tags = append(reg.Tags, opt.Tags...)

	check := new(api.AgentServiceCheck)
	check.HTTP = fmt.Sprintf("http://%s:%d/ping", reg.Address, reg.Port)
	check.Timeout = "5s"
	check.Interval = "3s"
	check.DeregisterCriticalServiceAfter = "300s"
	reg.Check = check
	for {
		if err := client.Agent().ServiceRegister(reg); err != nil {
			glog.Error(err)
			time.Sleep(time.Second)
		} else {
			glog.Info("Success Regitser Service Http", reg.ID)
			break
		}
	}
	//注册Apache APISIX ConsulKV服务发现
	PutUpstreamsToConsulKV(opt)
}

// apache APISIX ConsulKV 服务发现
func PutUpstreamsToConsulKV(opt *sys.Options) {
	consulConfig := api.DefaultConfig()
	consulConfig.Address = opt.ConsulAddress
	client, err := api.NewClient(consulConfig)
	if err != nil {
		glog.Error(err)
		return
	}

	// 删除旧K
	k := fmt.Sprintf("upstreams/%s/%s:", opt.Name, opt.Address)
	s, _, _ := client.KV().Keys(k, "", nil)
	for _, v := range s {
		for {
			if _, err := client.KV().Delete(v, nil); err != nil {
				glog.Error(v, "delete error ", err)
				time.Sleep(time.Second)
			} else {
				glog.Info("delete k ", v)
				break
			}
		}
	}

	// 加入新K
	kv := api.KVPair{
		Key:   fmt.Sprintf("upstreams/%s/%s:%d", opt.Name, opt.Address, *opt.HttpPort),
		Value: []byte(`{"weight": 1, "max_fails": 2, "fail_timeout": 3}`),
	}
	for {
		if _, err := client.KV().Put(&kv, nil); err != nil {
			glog.Error(err)
			time.Sleep(time.Second)
		} else {
			glog.Info("put new k ", kv.Key)
			break
		}
	}
}
