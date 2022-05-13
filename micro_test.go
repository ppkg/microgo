package microgo

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ppkg/microgo/sys"
)

func TestMicro_Run(t *testing.T) {
	var httpPort = 88

	Init(&sys.Options{
		Name:           "microgo",
		Path:           "test-path",
		JaegerEndpoint: "",
		XxljobAddress:  "",
		ConsulAddress:  "127.0.0.1:8500",
		HttpPort:       &httpPort,
	}).Run(routes)

}

func routes(r *gin.RouterGroup) {
	r.POST("/test", func(ctx *gin.Context) { ctx.String(200, "ok") })
	r.GET("/test", func(ctx *gin.Context) { ctx.String(200, "ok") })
}
