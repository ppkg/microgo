package server

import (
	"context"
	"fmt"
	"net"
	"runtime/debug"

	"github.com/ppkg/microgo/consul"
	"github.com/ppkg/microgo/grpc/gateway"
	"github.com/ppkg/microgo/sys"

	// grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"

	gwruntime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/maybgit/glog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

func Run(ctx context.Context, f func(*grpc.Server)) {
	opt := sys.GetOption()
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", *opt.GrpcPort))
	if err != nil {
		glog.Error(err)
	}
	defer func() {
		if err := l.Close(); err != nil {
			glog.Errorf("Failed to close tcp %d: %v", *opt.GrpcPort, err)
		}
	}()

	// 返回动态的端口
	if *opt.GrpcPort == 0 {
		*opt.GrpcPort = l.Addr().(*net.TCPAddr).Port
	}
	s := grpc.NewServer(grpc.UnaryInterceptor(filter))
	// s := grpc.NewServer(grpc.UnaryInterceptor(
	// 	grpc_auth.UnaryServerInterceptor(func(ctx context.Context) (context.Context, error) {
	// 		// return ctx,errors.New("not auth")
	// 		return ctx, nil
	// 	}),
	// ))

	f(s)

	grpc_health_v1.RegisterHealthServer(s, &consul.HealthImpl{})
	reflection.Register(s)
	go func() {
		defer s.GracefulStop()
		<-ctx.Done()
	}()
	if err := s.Serve(l); err != nil {
		glog.Error(err)
	}
}

func RunServerAndGateway(f func(s *grpc.Server), handler ...func(ctx context.Context, mux *gwruntime.ServeMux, conn *grpc.ClientConn) error) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go Run(ctx, f)
	go consul.RegisterGrpcService()
	gateway.Run(ctx, handler...)
}

func filter(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	defer func() {
		if e := recover(); e != nil {
			debug.PrintStack()
			err = status.Errorf(codes.Internal, "%v", e)
		}
	}()

	return handler(ctx, req)
}
