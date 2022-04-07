package microgo

import (
	"context"

	"github.com/gin-gonic/gin"
	gwruntime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	ginfw "github.com/ppkg/microgo/gin"
	"github.com/ppkg/microgo/sys"
	"github.com/ppkg/microgo/xxljob"
	"google.golang.org/grpc"
)

type Micro struct {}

func Init(opt *sys.Options) *Micro {
	sys.Set(opt)
	return &Micro{}
}

func (s *Micro) Run(fn interface{}, middleware ...interface{}) {
	go xxljob.Run()

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

	switch v := fn.(type) {
	case func(r *gin.RouterGroup):
		ginfw.Run(v, ginHandler...)
	case func(s *grpc.Server):
		runServerAndGateway(v, grpcHandler...)
	}
}
