# microgo
microservices for go
```
package main

import (
	"_/common/proto/pb/system" // you pb file package path
	"_/services"
	"flag"
	_ "net/http/pprof"

	"github.com/ppkg/microgo/consul"
	"github.com/ppkg/microgo/sys"

	"github.com/maybgit/glog"
	"google.golang.org/grpc"
)

func main() {
	flag.Parse()
	defer glog.Flush()

	sys.Init(&sys.Options{
		Name: "system-service",
	})

	consul.RegisterPprof()
	server.RunServerAndGateway(
		func(s *grpc.Server) {
			system.RegisterSystemServer(s, &services.System{})
		},
		system.RegisterSystemHandler,
	)
}

```
