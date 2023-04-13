package config

import (
	"gopkg.in/ini.v1"
)

var DefaultCfg = &Config{}
var defaultConfigPath = "./config.ini"

func LoadDefaultConfig(path string) error {
	var (
		pf  *ini.File
		err error
	)
	pf, err = ini.Load(path)
	if err != nil {
		return err
	}
	err = pf.MapTo(DefaultCfg)
	if err != nil {
		return err
	}
	return nil
}

// Kubernetes 与 kubernetes 相关的配置, 如果写了 KubeConfig 就不用再填其它的了
type Kubernetes struct {
	Namespace  string `ini:"namespace"`
	KubeConfig string `ini:"kubeConfig"`
	Token      string `ini:"token"`
	MasterUrl  string `ini:"masterUrl"`
}

// Envoy 与 envoy 相关配置, 添加一些默认选项
type Envoy struct {
}

type Config struct {
	Kubernetes `ini:"kubernetes"`
	Envoy      `ini:"envoy"`
}

// Load 只会在启动时被调用
func Load() error {
	err := LoadDefaultConfig(defaultConfigPath)
	if err != nil {
		return err
	}
	return nil
}

// GetConfig 获取配置使用这个方法
func GetConfig() Config {
	return *DefaultCfg
}
