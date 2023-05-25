package main

import (
	"github.com/wonderivan/logger"
	"github.com/xdmybl/engine-gate/app/grm"
	"github.com/xdmybl/engine-gate/app/xrm"
	"github.com/xdmybl/engine-gate/util/config"
	"github.com/xdmybl/engine-gate/util/constant"
	"github.com/xdmybl/engine-gate/util/log"
	"os"
	"os/signal"
	"syscall"
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
	logger.Info("config: %v", config.GetConfig())
	// init grm
	err = grm.InitDefaultGRM()
	if err != nil {
		logger.Error("error code: %v , err:  %v", constant.GRMError, err)
		os.Exit(constant.GRMError)
	}
	err = xrm.Init()
	if err != nil {
		logger.Error("error code: %v , err:  %v", constant.XRMError, err)
		os.Exit(constant.XRMError)
	}

	shutdownChannel := make(chan os.Signal, 1)

	// 启动 xds server 协程
	xrm.RunXDSServer(shutdownChannel)

	signalChan := make(chan os.Signal, 1)
	// `os.Interrupt` 信号通常是由用户在终端上按下 `Ctrl+C` 触发的。当用户按下 `Ctrl+C` 时，操作系统会向程序发送一个 `SIGINT` 信号，该信号会被转换为 `os.Interrupt` 信号。程序可以使用 `os/signal` 包中的 `os/signal.Notify` 函数来监听 `os.Interrupt` 信号，以便在接收到该信号时执行一些清理操作并优雅地退出。
	//`syscall.SIGTERM` 信号通常是由操作系统或其他进程发送的。例如，当用户在终端上使用 `kill` 命令时，操作系统会向指定的进程发送 `SIGTERM` 信号。程序可以使用 `os/signal` 包中的 `os/signal.Notify` 函数来监听 `syscall.SIGTERM` 信号，以便在接收到该信号时执行一些清理操作并优雅地退出。
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	// 等待信号
	<-signalChan

	// 停止服务器
	shutdownChannel <- syscall.SIGTERM

	// 其它协程都优雅退出后, 主协程退出
	_ = <-shutdownChannel
}
