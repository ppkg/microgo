package microgo

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"runtime/debug"
	"time"

	"github.com/ppkg/microgo/consul"
	"github.com/ppkg/microgo/sys"
	"github.com/ppkg/microgo/utils"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/maybgit/glog"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func runGin(route func(r *gin.RouterGroup), middleware ...gin.HandlerFunc) {
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
	ge.Use(recoverGin)
	ge.Use(preHandler())
	for _, v := range middleware {
		ge.Use(v)
	}
	ge.Use(logger())

	r := ge.Group("/" + opt.Path)
	route(r)

	ge.GET(fmt.Sprintf("/%s/swagger/*any", opt.Path), ginSwagger.WrapHandler(swaggerFiles.Handler))
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

// 中间件忽略日志的路径
var excludeLogPath = map[string]string{
	"/ping":                                  "",
	"/message-api/im/callback-after-message": "",
}

func preHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "POST" {
			b, _ := ioutil.ReadAll(c.Request.Body)
			c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(b))
			c.Set("BodyBytes", b)
		}
	}
}

func logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		if _, has := excludeLogPath[c.Request.URL.Path]; !has {
			start := time.Now()
			c.Next()
			end := time.Now()
			loginName, _ := c.Get("LoginName")
			var bodyString = ""
			if c.Request.Method == "POST" {
				if v, has := c.Get("BodyBytes"); has {
					bodyString = string(v.([]byte))
				}
			}
			glog.Info(c.Request.Header, loginName, c.Request.Method, c.Request.URL.Path, c.Writer.Status(), end.Sub(start), bodyString)
		}
	}
}

func recoverGin(c *gin.Context) {
	defer func() {
		if err := recover(); err != nil {
			glog.Error(err)
			glog.Error(string(debug.Stack()))
			errStr := utils.ErrorToString(err)
			c.JSON(400, gin.H{
				"message": errStr,
			})
			c.Abort()
		}
	}()
	c.Next()
}
