package microgo

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/ppkg/microgo/consul"
	"github.com/ppkg/microgo/sys"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/maybgit/glog"
	_ "google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/metadata"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"

	gwruntime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

type grpcServer struct{}

func runGateway(ctx context.Context, handler ...func(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error) {
	opts := sys.GetOption()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	conn, f := dial(ctx, opts.GrpcPort)
	defer f()

	mux := http.NewServeMux()
	mux.HandleFunc("/swagger/", serveSwaggerUi)
	mux.HandleFunc("/ping", func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Content-Type", "text/plain")
		if s := conn.GetState(); s != connectivity.Ready {
			http.Error(rw, fmt.Sprintf("grpc server is %s", s), http.StatusBadGateway)
			return
		}
		fmt.Fprint(rw, opts.Name)
	})

	gw := runtime.NewServeMux(opts.Mux...)
	for _, fn := range handler {
		if err := fn(ctx, gw, conn); err != nil {
			glog.Error(err)
		}
	}

	mux.Handle("/", gw)

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", *opts.HttpPort))
	if err != nil {
		glog.Error(err)
		panic(l)
	}

	// 返回动态端口
	if *opts.HttpPort == 0 {
		*opts.HttpPort = l.Addr().(*net.TCPAddr).Port
	}

	opt := *opts
	opt.Name += "-gw"
	opt.Tags = opt.Tags[0:0]
	consul.RegisterHttpService(&opt)
	if err := http.Serve(l, allowCORS(mux)); err != nil {
		glog.Error(err)
		panic(err)
	}
}

func serveSwaggerUi(w http.ResponseWriter, r *http.Request) {
	p := strings.TrimPrefix(r.URL.Path, "/swagger/")
	p = path.Join("static/swagger-ui/", p)
	http.ServeFile(w, r, p)
}

func allowCORS(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if origin := r.Header.Get("Origin"); origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			if r.Method == "OPTIONS" && r.Header.Get("Access-Control-Request-Method") != "" {
				g.preflightHandler(w, r)
				return
			}
		}
		h.ServeHTTP(w, r)
	})
}

func preflightHandler(w http.ResponseWriter, r *http.Request) {
	headers := []string{"Content-Type", "Accept", "Authorization"}
	w.Header().Set("Access-Control-Allow-Headers", strings.Join(headers, ","))
	methods := []string{"GET", "HEAD", "POST", "PUT", "DELETE"}
	w.Header().Set("Access-Control-Allow-Methods", strings.Join(methods, ","))
	glog.Infof("preflight request for %s", r.URL.Path)
}

func dial(ctx context.Context, grpc_port *int) (*grpc.ClientConn, func()) {
	for *grpc_port == 0 {
		time.Sleep(time.Second * 1)
	}
	conn, err := grpc.Dial(fmt.Sprintf(":%d", *grpc_port), grpc.WithInsecure())
	if err != nil {
		glog.Error(err)
		return nil, nil
	}
	return conn, func() {
		if err != nil {
			if cerr := conn.Close(); cerr != nil {
				glog.Infof("Failed to close conn to %s: %v", *grpc_port, cerr)
			}
			return
		}
		go func() {
			<-ctx.Done()
			if cerr := conn.Close(); cerr != nil {
				glog.Infof("Failed to close conn to %s: %v", *grpc_port, cerr)
			}
		}()
	}
}

func runServer(ctx context.Context, f func(*grpc.Server)) {
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
	s := grpc.NewServer(
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_recovery.UnaryServerInterceptor(),
			otelgrpc.UnaryServerInterceptor(),
		)),
		grpc.StreamInterceptor(otelgrpc.StreamServerInterceptor()),
	)

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

func runServerAndGateway(f func(s *grpc.Server), handler ...func(ctx context.Context, mux *gwruntime.ServeMux, conn *grpc.ClientConn) error) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go runServer(ctx, f)
	go consul.RegisterGrpcService()
	runGateway(ctx, handler...)
}



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
