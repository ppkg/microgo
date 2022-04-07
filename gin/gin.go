package gin

import (
	"fmt"
	"net"
	"time"

	"github.com/ppkg/microgo/consul"
	"github.com/ppkg/microgo/sys"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/maybgit/glog"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func Run(route func(r *gin.RouterGroup), middleware ...gin.HandlerFunc) {
	opt := sys.GetOption()
	ge := gin.New()

	gin.DebugPrintRouteFunc = func(httpMethod, absolutePath, handlerName string, nuHandlers int) {
		glog.Infof("%v %v %v %v", httpMethod, absolutePath, handlerName, nuHandlers)
	}

	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowHeaders = append(config.AllowHeaders, "Authorization")
	sys.UseTrace(func() {
		ge.Use(otelgin.Middleware(opt.Name))
	})
	ge.Use(cors.New(config))
	ge.Use(Recover)
	ge.Use(PreHandler())
	for _, v := range middleware {
		ge.Use(v)
	}
	ge.Use(Logger())

	r := ge.Group("/" + opt.Name)
	route(r)

	ge.GET(fmt.Sprintf("/%s/swagger/*any", opt.Name), ginSwagger.WrapHandler(swaggerFiles.Handler))
	ge.GET("/ping", func(c *gin.Context) {
		c.String(200, opt.Name)
	})
	if l, err := net.Listen("tcp", fmt.Sprintf(":%d", *opt.HttpPort)); err != nil {
		glog.Error(err)
	} else {
		go func() {
			for *opt.HttpPort == 0 {
				*opt.HttpPort = l.Addr().(*net.TCPAddr).Port
				time.Sleep(time.Second)
			}

			// 注册所有集群
			consul.RegisterHttpService(opt)

			// 注册为每个单机的模式，用于灰度控制
			o := *opt
			o.Name += "-" + o.LocalIP
			consul.RegisterHttpService(&o)
		}()

		err := ge.RunListener(l)
		if err != nil {
			glog.Error("gin start error", err)
			panic(err)
		}
	}
}
