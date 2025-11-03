package monitor

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/khiemnd777/andy_api/shared/logger"
)

func InitModuleLifecycle(name string, port int) {
	defer ModulePanic(name, port)
	SetModuleStatus(name, port, "RUNNING")

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-c
		logger.Warn("📴 Graceful shutdown: " + name)
		SetModuleStatus(name, port, "STOPPED")
		os.Exit(0)
	}()
}

func ModulePanic(name string, port int) {
	if r := recover(); r != nil {
		logger.Error(fmt.Sprintf("🔥 Panic in module [%s]: %v", name, r))
		SetModuleStatus(name, port, "FAILED")
		os.Exit(1)
	}
}
