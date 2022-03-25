package xxljob

import "github.com/maybgit/glog"

// LogFunc 应用日志
type LogFunc func(req LogReq, res *LogRes) []byte

// Logger 系统日志
type Logger interface {
	Infof(format string, a ...interface{})
	Errorf(format string, a ...interface{})
	Info(a ...interface{})
	Error(a ...interface{})
}

type logger struct {
}

func (l *logger) Info(a ...interface{}) {
	glog.Info(a...)
}

func (l *logger) Error(a ...interface{}) {
	glog.Error(a...)
}

func (l *logger) Infof(format string, a ...interface{}) {
	glog.Infof(format, a...)
}

func (l *logger) Errorf(format string, a ...interface{}) {
	glog.Errorf(format, a...)
}
