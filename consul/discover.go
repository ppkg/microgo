package consul

import (
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/ppkg/microgo/sys"

	consulapi "github.com/hashicorp/consul/api"
	"github.com/maybgit/glog"
	"google.golang.org/grpc/resolver"
)

var (
	services  sync.Map
	lw        sync.RWMutex           // lock watcher
	watcherFn = make(map[string]int) // 保存服务watcher，结合lock watcher，保证每个服务只有一个watcher
)

func init() {
	resolver.Register(NewBuilder())

	//监测当前watcherFn，可以此协程可以随时注释
	go func() {
		for {
			for k := range watcherFn {
				glog.Info("watcherFn", k)
			}
			time.Sleep(time.Second * 30)
		}
	}()
}

type consulResolver struct {
	cc          resolver.ClientConn
	serviceName string
	consulAddr  string
}

func newConsulResolver(cc resolver.ClientConn, scheme, consulAddr, serviceName string) (*consulResolver, error) {
	return &consulResolver{
		cc:          cc,
		serviceName: serviceName,
		consulAddr:  consulAddr,
	}, nil
}

func (c *consulResolver) start() {
	getService := func() bool {
		if v, has := services.Load(c.serviceName); has {
			rs := v.(resolver.State)
			for _, v := range rs.Addresses {
				glog.Info(v.ServerName, v.Addr)
			}
			// 灰度发布处理，如果状态为灰度，服务只取本机服务
			// TODO 如果同一套服务没有在一台机上部署完整，此方式有问题
			// 因此灰度的方式得改变，这里用简单的在一台机上部署所有依赖的服务
			// 后期有需要不同的服务，需要在不同的机器上部署再做处理
			if sys.GlobalConfig.HdRelease {
				glog.Info("sys.GlobalConfig.HdRelease", sys.GlobalConfig.HdRelease)
				cfg := sys.GetOption()
				for _, v := range rs.Addresses {
					if strings.HasPrefix(v.Addr, cfg.Address) {
						c.cc.UpdateState(resolver.State{Addresses: []resolver.Address{{Addr: v.Addr, ServerName: v.ServerName}}})
						break
					}
				}
			} else {
				c.cc.UpdateState(rs)
			}
			return true
		} else {
			return false
		}
	}

	//是否获取到服务
	if !getService() {
		lw.Lock()
		defer lw.Unlock()
		// 一个服务只有一个watcher
		if _, has := watcherFn[c.serviceName]; !has {
			watcherFn[c.serviceName] = 1
			go watcher(c.serviceName, c.consulAddr)
		}

		for !getService() {
			glog.Info(" wait watcher...", c.serviceName)
			time.Sleep(time.Second)
		}
	}
}

func watcher(serviceName, consulAddr string) {
	var (
		tags        = make(map[string]string)
		lastIndex   uint64
		reqErrCount int // 从consul服务器没有获取到服务的次数
	)
	// 服务发现只查有grpc tag的服务
	tags["grpc"] = "grpc"

	for {
		// 以闭包方式执行，recover有错的保证协程不退出
		func() {
			defer func() {
				if err := recover(); err != nil {
					glog.Error(err)
					time.Sleep(time.Second * 3)
				}
			}()

			glog.Info("consul watcher lastIndex", lastIndex)

			var newTags []string
			for _, v := range tags {
				if v != "" {
					newTags = append(newTags, v)
				}
			}

			glog.Info("service tags ", newTags)
			client, err := consulapi.NewClient(&consulapi.Config{Address: consulAddr, Scheme: "http"})
			if err != nil {
				glog.Info(err)
			}

			entry, meta, err := client.Health().ServiceMultipleTags(serviceName, newTags, true, &consulapi.QueryOptions{WaitIndex: lastIndex})
			if err != nil {
				// 屏蔽SLB超时的错误打印
				// glog.Error(err)
				glog.Info("watch break off", err)
			} else {
				glog.Info("entry len", len(entry))

				lastIndex = meta.LastIndex
				var sta resolver.State
				for _, v := range entry {
					// 兼容本地 network host Docker Desktop 通信
					if runtime.GOOS == "windows" && strings.HasPrefix(v.Service.Address, "172.") {
						v.Service.Address = "127.0.0.1"
					}
					sta.Addresses = append(sta.Addresses, resolver.Address{
						Addr:       fmt.Sprintf("%s:%d", v.Service.Address, v.Service.Port),
						ServerName: v.Service.Service,
					})
				}

				c := len(sta.Addresses)
				if c == 0 {
					reqErrCount += 1
					if reqErrCount > 5 {
						services.Delete(serviceName)
					}
				} else {
					reqErrCount = 0
					services.Store(serviceName, sta)
					for _, v := range sta.Addresses {
						glog.Info(v.ServerName, v.Addr)
					}
				}
			}
		}()
	}
}

func (c *consulResolver) ResolveNow(o resolver.ResolveNowOptions) {}

func (c *consulResolver) Close() {}
