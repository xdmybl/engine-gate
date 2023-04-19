package config

import (
	"gopkg.in/ini.v1"
)

var DefaultCfg = &Config{}
var defaultConfigPath = "./config.ini"

// TODO 设置各种选项默认值

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
	KubeConfig string `ini:"kube_config"`
	Token      string `ini:"token"`
	MasterUrl  string `ini:"master_url"`
}

type XDSOption struct {
}

type Envoy struct {
	NodeId string `ini:"node_id"`
}

// XDS 与 envoy 相关配置, 添加一些默认选项
type XDS struct {
	ListenIp string `ini:"listen_ip"`
	Port     string `ini:"port"`
}

type Config struct {
	Kubernetes `ini:"kubernetes"`
	XDS        `ini:"xds"`
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
