package xxljob

import (
	"context"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/maybgit/glog"
)

// TaskFunc 任务执行函数
type TaskFunc func(cxt context.Context, param *RunReq) string

// Task 任务
type Task struct {
	Id        int64
	Name      string
	Ctx       context.Context
	Param     *RunReq
	fn        TaskFunc
	Cf        context.CancelFunc
	StartTime time.Time
	log       Logger
}

// 任务耗时
func (t *Task) TimeConsuming() int64 {
	return time.Now().Unix() - t.StartTime.Unix()
}

func (t *Task) FormatStartTime() string {
	return t.StartTime.Format("2006-01-02 15:04:05")
}

// Run 运行任务
func (t *Task) Run(logID int64, callback func(code int64, msg string)) {
	defer func(cf func()) {
		if err := recover(); err != nil {
			if t.Param.LogID == logID {
				glog.Error("任务ID["+Int64ToStr(t.Id)+"]任务名称["+t.Name+"]参数："+t.Param.ExecutorParams, err)
				debug.PrintStack() //堆栈跟踪
				callback(500, "task panic："+fmt.Sprintf("%v", err))
				if cf != nil {
					cf()
				}
			}
		}
	}(t.Cf)
	msg := t.fn(t.Ctx, t.Param)
	// glog.Info(t.Param.LogID, logID)
	if t.Param.LogID == logID {
		callback(200, fmt.Sprintf("%s 任务执行完成，开始时间：%s，耗时：%ds", msg, t.FormatStartTime(), t.TimeConsuming()))
	}
}
