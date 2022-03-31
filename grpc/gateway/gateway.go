package gateway

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
)

func Run(ctx context.Context, handler ...func(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error) {
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
				preflightHandler(w, r)
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
