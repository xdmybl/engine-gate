package main

import (
	"github.com/wonderivan/logger"
	"github.com/xdmybl/engine-gate/app/grm"
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
		logger.Error("error code: %v , err:  %v", constant.ConfigError, err)
		os.Exit(constant.ConfigError)
	}
	// init grm
	err = grm.InitDefaultGRM()
	if err != nil {
		logger.Error("error code: %v , err:  %v", constant.GRMError, err)
	}
}
