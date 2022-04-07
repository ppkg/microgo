package xxljob

import (
	"context"
	"sync"

	"github.com/ppkg/microgo/conn"
	"github.com/ppkg/microgo/sys"

	"github.com/maybgit/glog"
)

var (
	exec Executor
	jobs = make(map[string]func(cxt context.Context, param *RunReq) (msg string))
	lock sync.RWMutex
)

// 添加标准的xxljob任务
// jobHandler：同一appid下不能重复
func AddTask(jobName string, jobFunc func(ctx context.Context, param *RunReq) (msg string)) {
	lock.Lock()
	defer lock.Unlock()
	if _, has := jobs[jobName]; has {
		glog.Error(jobName, "duplicate definition")
	} else {
		jobs[jobName] = func(cxt context.Context, param *RunReq) (msg string) {
			cxt, _ = conn.GetCtx(cxt)
			return jobFunc(cxt, param)
		}
		glog.Info("xxljob.AddTask", jobName)
	}
}

func Run() {
	opt := sys.GetOption()

	if opt.XxljobAddress == "" {
		glog.Error("xxljob server address is empty")
	} else {
		exec = NewExecutor(
			ServerAddr(opt.XxljobAddress),
			ExecutorPort(opt.XxljobPort),
			RegistryKey(opt.Name),
			SetLogger(&logger{}),
		)
		exec.Init()
		glog.Info(opt.Name, "jobs", len(jobs))
		for k, v := range jobs {
			glog.Info("RegTask：", k)
			exec.RegTask(k, v)
		}
		if err := exec.Run(); err != nil {
			glog.Info("xxljob exec run error ", err)
		}
	}
}
