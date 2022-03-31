package utils

import (
	"github.com/maybgit/glog"
	"google.golang.org/grpc"
)

func ErrorToString(err interface{}) (str string) {
	str = "utils.ErrorToString 类型断言失败"
	defer func() {
		if err := recover(); err != nil {
			glog.Error(err)
		}
	}()
	switch v := err.(type) {
	case error:
		// str = v.Error()
		str = grpc.ErrorDesc(v)
	default:
		str = err.(string)
	}
	return
}
