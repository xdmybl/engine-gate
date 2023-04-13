package main

import (
	"github.com/wonderivan/logger"
	"github.com/xdmybl/engine-gate/util/config"
	"github.com/xdmybl/engine-gate/util/constant"
	"github.com/xdmybl/engine-gate/util/log"
	"os"
)

func main() {
	// server init
	log.InitLogger()
	// load config
	err := config.Load()
	if err != nil {
		logger.Error(constant.ConfigError)
		os.Exit(constant.ConfigError)
	}
	// init grm
}
