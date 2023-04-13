package grm

import (
	"github.com/wonderivan/logger"
)

// GRM (gate resource manager) 负责 gate resource 的管理
// 里面至少要保存 具体类型的 对象管理器, 每个管理器中包括相应的 对象客户端 , (why? 因为 kubernetes 可能提供不了我们想要的复杂查询时, 就需要
// 在 管理器这一层提供复杂查询接口)
type GRM struct {
	CaManager CaManager
}

func (g *GRM) Init() error {
	var cm CaManager
	err := cm.Init()
	if err != nil {
		logger.Error("ca client init err:  %v", err)
		return err
	}
	return nil
}
