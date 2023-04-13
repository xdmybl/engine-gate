package log

import (
	"github.com/wonderivan/logger"
)

func InitLogger() {
	logger.SetLogger("./log.json")
	logger.Debug("debug init : %v ", logger.LevelMap)
}
