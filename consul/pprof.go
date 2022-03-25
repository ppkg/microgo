package consul

import (
	"fmt"
	"net"
	"net/http"

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
		RegisterHttpService(&sys.Options{
			ConsulAddress: opt.ConsulAddress,
			Address:       opt.Address,
			Name:          name,
			Tags:          []string{"pprof"},
			HttpPort:      new(int),
			PprofPort:     new(int),
		}, l)
		RegisterHttpService(&sys.Options{
			ConsulAddress: opt.ConsulAddress,
			Address:       opt.Address,
			Name:          fmt.Sprintf("%s-pprof-%s", opt.Name, opt.Address),
			Tags:          []string{"pprof"},
			HttpPort:      new(int),
			PprofPort:     new(int),
		}, l)
		http.HandleFunc("/ping", func(rw http.ResponseWriter, r *http.Request) {
			rw.Write([]byte(name))
		})

		// *opt.PprofPort = l.Addr().(*net.TCPAddr).Port
		go http.Serve(l, nil)
	}
}
