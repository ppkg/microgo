package conn

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/maybgit/glog"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
)

var ErrRecordNotFound = errors.New("record not found")

const (
	Silent logger.LogLevel = iota + 1
	Error
	Warn
	Info
)

// Writer log writer interface
type Writer interface {
	Printf(string, ...interface{})
}

func New(writer Writer, config logger.Config) logger.Interface {
	var (
		infoStr      = "%s "
		warnStr      = "%s "
		errStr       = "%s "
		traceStr     = "%s [%.3fms] [rows:%v] %s"
		traceWarnStr = "%s %s [%.3fms] [rows:%v] %s"
		traceErrStr  = "%s %s [%.3fms] [rows:%v] %s"
	)

	return &gormlogger{
		Writer:       writer,
		Config:       config,
		infoStr:      infoStr,
		warnStr:      warnStr,
		errStr:       errStr,
		traceStr:     traceStr,
		traceWarnStr: traceWarnStr,
		traceErrStr:  traceErrStr,
	}
}

type gormlogger struct {
	Writer
	logger.Config
	infoStr, warnStr, errStr            string
	traceStr, traceErrStr, traceWarnStr string
}

// LogMode log mode
func (l *gormlogger) LogMode(level logger.LogLevel) logger.Interface {
	newlogger := *l
	newlogger.LogLevel = level
	return &newlogger
}

// Info print info
func (l gormlogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= Info {
		glog.Infof(l.infoStr+msg, append([]interface{}{utils.FileWithLineNum()}, data...)...)
	}
}

// Warn print warn messages
func (l gormlogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= Warn {
		glog.Warningf(l.warnStr+msg, append([]interface{}{utils.FileWithLineNum()}, data...)...)
	}
}

// Error print error messages
func (l gormlogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= Error {
		glog.Errorf(l.errStr+msg, append([]interface{}{utils.FileWithLineNum()}, data...)...)
	}
}

// Trace print sql message
func (l gormlogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.LogLevel <= Silent {
		return
	}

	elapsed := time.Since(begin)
	switch {
	case err != nil && l.LogLevel >= Error && (!errors.Is(err, ErrRecordNotFound) || !l.IgnoreRecordNotFoundError):
		sql, rows := fc()
		if rows == -1 {
			glog.Errorf(l.traceErrStr, utils.FileWithLineNum(), err, float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			glog.Errorf(l.traceErrStr, utils.FileWithLineNum(), err, float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	case elapsed > l.SlowThreshold && l.SlowThreshold != 0 && l.LogLevel >= Warn:
		sql, rows := fc()
		slowLog := fmt.Sprintf("SLOW SQL >= %v", l.SlowThreshold)
		if rows == -1 {
			glog.Warningf(l.traceWarnStr, utils.FileWithLineNum(), slowLog, float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			glog.Warningf(l.traceWarnStr, utils.FileWithLineNum(), slowLog, float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	case l.LogLevel == Info:
		sql, rows := fc()
		if rows == -1 {
			glog.Infof(l.traceStr, utils.FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			glog.Infof(l.traceStr, utils.FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	}
}
