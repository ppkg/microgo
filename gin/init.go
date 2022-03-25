package gin

import (
	"fmt"
	"net"
	"time"

	"github.com/ppkg/microgo/consul"
	"github.com/ppkg/microgo/sys"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/maybgit/glog"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type engine struct {
	ge  *gin.Engine
	opt *sys.Options
}

func (e *engine) Start() {
	e.ge.GET(fmt.Sprintf("/%s/swagger/*any", e.opt.Name), ginSwagger.WrapHandler(swaggerFiles.Handler))
	e.ge.GET("/ping", func(c *gin.Context) {
		c.String(200, e.opt.Name)
	})
	if l, err := net.Listen("tcp", fmt.Sprintf(":%d", *e.opt.HttpPort)); err != nil {
		glog.Error(err)
	} else {
		go func() {
			for *e.opt.HttpPort == 0 {
				*e.opt.HttpPort = l.Addr().(*net.TCPAddr).Port
				time.Sleep(time.Second)
			}

			// 注册所有集群
			consul.RegisterHttpService(e.opt)

			// 注册为每个单机的模式，用于灰度控制
			o := *e.opt
			o.Name += "-" + o.Address
			consul.RegisterHttpService(&o)
		}()
		
		err := e.ge.RunListener(l)
		if err != nil {
			glog.Error("gin start error", err)
			panic(err)
		}
	}
}

func (e *engine) Routes(route ...func(r *gin.RouterGroup)) *engine {
	r := e.ge.Group("/" + e.opt.Name)
	for _, f := range route {
		f(r)
	}
	return e
}

func (e *engine) Middlewares(middleware ...gin.HandlerFunc) *engine {
	for _, v := range middleware {
		e.ge.Use(v)
	}
	e.ge.Use(Logger())
	return e
}

func Init() *engine {
	opt := sys.GetOption()
	gin.DebugPrintRouteFunc = func(httpMethod, absolutePath, handlerName string, nuHandlers int) {
		glog.Infof("%v %v %v %v", httpMethod, absolutePath, handlerName, nuHandlers)
	}

	e := engine{
		ge:  gin.New(),
		opt: opt,
	}
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowHeaders = append(config.AllowHeaders, "Authorization")
	e.ge.Use(cors.New(config))
	e.ge.Use(Recover)
	e.ge.Use(PreHandler())
	return &e
}
