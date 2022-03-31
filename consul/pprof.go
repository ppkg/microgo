package consul

import (
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/ppkg/microgo/sys"

	"github.com/maybgit/glog"
)

// 启动 pprof性能分析，把端口注册到consul
func RegisterPprof() {
	opt := sys.GetOption()
	if l, err := net.Listen("tcp", fmt.Sprintf(":%d", *opt.PprofPort)); err != nil {
		glog.Error(err)
	} else {
		name := fmt.Sprintf("%s-pprof", opt.Name)

		go func() {
			for *opt.PprofPort == 0 {
				*opt.PprofPort = l.Addr().(*net.TCPAddr).Port
				glog.Info("pprof port ", *opt.PprofPort)
				time.Sleep(time.Second)
			}

			// 注册集群模式
			RegisterHttpService(&sys.Options{
				ConsulAddress: opt.ConsulAddress,
				LocalIP:       opt.LocalIP,
				Name:          name,
				Tags:          []string{"pprof"},
				HttpPort:      new(int),
				PprofPort:     opt.PprofPort,
			})

			// 注册单机模式
			RegisterHttpService(&sys.Options{
				ConsulAddress: opt.ConsulAddress,
				LocalIP:       opt.LocalIP,
				Name:          fmt.Sprintf("%s-pprof-%s", opt.Name, opt.LocalIP),
				Tags:          []string{"pprof"},
				HttpPort:      new(int),
				PprofPort:     opt.PprofPort,
			})
		}()

		http.HandleFunc("/ping", func(rw http.ResponseWriter, r *http.Request) {
			rw.Write([]byte(name))
		})
		go http.Serve(l, nil)
	}
}
