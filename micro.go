package microgo

import (
	"context"
	"strings"

	"github.com/gin-gonic/gin"
	gwruntime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/ppkg/microgo/consul"
	"github.com/ppkg/microgo/sys"
	"github.com/ppkg/microgo/xxljob"
	"google.golang.org/grpc"
)

type Micro struct {
	opt *sys.Options
}

func Init(o *sys.Options) *Micro {
	sys.Init(o)
	o.Path = strings.Trim(o.Path, "/")
	if o.Path == "" {
		o.Path = o.Name
	}
	return &Micro{opt: o}
}

func (s *Micro) Run(fn interface{}, middleware ...interface{}) {
	var (
		ginHandler  []gin.HandlerFunc
		grpcHandler []func(ctx context.Context, mux *gwruntime.ServeMux, conn *grpc.ClientConn) error
	)

	for _, i := range middleware {
		switch v := i.(type) {
		case gin.HandlerFunc:
			ginHandler = append(ginHandler, v)
		case func(ctx context.Context, mux *gwruntime.ServeMux, conn *grpc.ClientConn) error:
			grpcHandler = append(grpcHandler, v)
		}
	}

	if s.opt.XxljobAddress != "" {
		go xxljob.Run()
	}

	if !s.opt.PprofDisable {
		consul.RegisterPprof()
	}

	switch v := fn.(type) {
	case func(r *gin.RouterGroup):
		runGin(v, ginHandler...)
	case func(s *grpc.Server):
		runServerAndGateway(v, grpcHandler...)
	}
}
