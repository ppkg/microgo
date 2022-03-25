package consul

import (
	"fmt"
	"log"
	"net"
	"testing"

	"github.com/ppkg/microgo/sys"
	"github.com/ppkg/microgo/utils"

	"github.com/hashicorp/consul/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
)

func TestRegitserService(t *testing.T) {
	lis, err := net.Listen("tcp", ":10001")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer(
	// grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
	// 	PermitWithoutStream: true,
	// }),
	)
	ip := utils.GetIp()
	fmt.Println(ip)
	RegisterGrpcService(&sys.Options{
		ConsulAddress: "127.0.0.1:8500",
		Name:          "logservice",
		Tag:           []string{"logservice"},
		// HttpPort:      10001,
	})
	grpc_health_v1.RegisterHealthServer(s, &HealthImpl{})

	// proto.RegisterLogServer(s, &services.LogService{})

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func TestKV(t *testing.T) {
	consulConfig := api.DefaultConfig()
	consulConfig.Address = "127.0.0.1:8500"
	client, err := api.NewClient(consulConfig)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	ks, _, _ := client.KV().Keys("upstreams/doctor-api/192.168.65.3:", "", nil)
	for _, v := range ks {
		t.Log(v)
	}
}
