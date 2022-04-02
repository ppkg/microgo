package sys

import (
	"encoding/json"
	"time"

	consulapi "github.com/hashicorp/consul/api"
	"github.com/maybgit/glog"
)

type SystemGlobalConfig struct {
	HdRelease bool //是否灰度
}

func init() {
	return //TODO
	// watch 服务配置
	go func() {
		var lastIndex uint64
		for {
			opt := GetOption()
			if client, err := consulapi.NewClient(&consulapi.Config{Address: opt.ConsulAddress, Scheme: "http"}); err != nil {
				glog.Warning(err)
			} else {
				pair, meta, err := client.KV().Get("SystemGlobalConfig", &consulapi.QueryOptions{WaitIndex: lastIndex})
				if err != nil {
					glog.Error(err)
				} else {
					if pair != nil {
						lastIndex = meta.LastIndex
						glog.Info("SystemGlobalConfig", string(pair.Value))
						json.Unmarshal(pair.Value, &GlobalConfig)
						glog.Info("GlobalConfig.HdRelease", GlobalConfig.HdRelease)
					} else {
						glog.Error("SystemGlobalConfig pair nil")
					}
				}
			}
			time.Sleep(time.Second * 3)
		}
	}()
}
