package conn

import (
	"context"
	"fmt"

	"github.com/maybgit/glog"
	"github.com/ppkg/microgo/sys"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type ConnClient struct {
	Conn *grpc.ClientConn
	Ctx  context.Context
	Cf   context.CancelFunc
}

var consulAddress = ""

func formatTarget(appid string) string {
	if consulAddress == "" {
		consulAddress = sys.GetOption().ConsulAddress
	}
	return fmt.Sprintf("consul://%s/%s", consulAddress, appid)
}

func GetConn(ctx context.Context, appid string) *grpc.ClientConn {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if _, has := md["X-Request-Id"]; !has {
			glog.Warning("X-Request-Id is empty")
		}
	}

	conn, err := grpc.DialContext(ctx, formatTarget(appid), grpc.WithBlock(), grpc.WithInsecure(), grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`))
	if err != nil {
		glog.Error(appid, err.Error())
	}
	return conn
}
