package grm

import (
	"github.com/wonderivan/logger"
)

var defaultGRM *GRM

// GetGRM 这个方法获取 grm 对象
func GetGRM() *GRM {
	return defaultGRM
}

func InitDefaultGRM() error {
	var err error
	defaultGRM, err = NewGRM()
	if err != nil {
		logger.Error("init default GRM err:  %v", err)
		return err
	}
	return nil
}

func NewGRM() (*GRM, error) {
	grm := new(GRM)
	err := grm.Init()
	if err != nil {
		return nil, err
	}
	return grm, nil
}

// GRM (gate resource manager) 负责 gate resource 的管理
// 里面至少要保存 具体类型的 对象管理器, 每个管理器中包括相应的 对象客户端 , (why? 因为 kubernetes 可能提供不了我们想要的复杂查询时, 就需要
// 在 管理器这一层提供复杂查询接口)
type GRM struct {
	CaManager       CaManager       `json:"ca_manager"`
	CertManager     CertManager     `json:"cert_manager"`
	UpstreamManager UpstreamManager `json:"upstream_manager"`
	FilterManager   FilterManager   `json:"filter_manager"`
	GatewayManager  GatewayManager  `json:"gateway_manager"`
}

func (g *GRM) Init() error {
	var cm CaManager
	err := cm.Init()
	if err != nil {
		logger.Error("ca client init err:  %v", err)
		return err
	}
	var cert CertManager
	err = cert.Init()
	if err != nil {
		logger.Error("cert client init err:  %v", err)
		return err
	}
	var upstreamManager UpstreamManager
	err = upstreamManager.Init()
	if err != nil {
		logger.Error("upstream client init err:  %v", err)
		return err
	}
	var filterManager FilterManager
	err = filterManager.Init()
	if err != nil {
		logger.Error("filter client init err:  %v", err)
		return err
	}
	var gatewayManager GatewayManager
	err = gatewayManager.Init()
	if err != nil {
		logger.Error("gateway client init err:  %v", err)
		return err
	}
	return nil
}
