package gin

import (
	"bytes"
	"io/ioutil"
	"runtime/debug"
	"time"

	"github.com/ppkg/microgo/utils"

	"github.com/gin-gonic/gin"
	"github.com/maybgit/glog"
)

// 中间件忽略日志的路径
var excludeLogPath = map[string]string{
	"/ping":                                  "",
	"/message-api/im/callback-after-message": "",
}

func PreHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "POST" {
			b, _ := ioutil.ReadAll(c.Request.Body)
			c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(b))
			c.Set("BodyBytes", b)
		}
	}
}

func Logger() gin.HandlerFunc {
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

func Recover(c *gin.Context) {
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
