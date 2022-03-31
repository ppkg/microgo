package utils

import (
	"github.com/maybgit/glog"
	"google.golang.org/grpc/status"
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
		str = status.Convert(v).Message()
	default:
		str = err.(string)
	}
	return
}
