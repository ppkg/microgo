package sys

import "flag"

var (
	grpcPort      = flag.Int("grpc_port", 0, "grpc服务的端口")
	httpPort      = flag.Int("http_port", 0, "http grpc-gateway的端口")
	pprofPort     = flag.Int("pprof_port", 0, "pprof的端口")
	consulAddress = flag.String("consul_address", "", "consul server地址")
	xxljobAddress = flag.String("xxljob_address", "", "xxljob后台地址，例如：http://xx.xx.xx.xx:8080/xxl-job-admin")
	xxljobPort    = flag.Int("xxljob_port", 0, "xxljob执行器端口")
)
