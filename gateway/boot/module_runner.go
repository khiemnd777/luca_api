package boot

import (
	"fmt"
	"net"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/khiemnd777/andy_api/shared/logger"
	"github.com/khiemnd777/andy_api/shared/utils"
)

func IsPortOpen(port int) bool {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), 1*time.Second)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

func WaitForModuleReady(host string, port int, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if utils.CheckPortOpen(host, port) {
			return true
		}
		time.Sleep(500 * time.Millisecond)
	}
	return false
}

func TryStartModule(moduleName string, port int) {
	if IsPortOpen(port) {
		logger.Info(fmt.Sprintf("Module [%s] already running", moduleName))
		return
	}

	go func() {
		mainFile := filepath.Join("modules", moduleName, "main.go")
		cmd := exec.Command("go", "run", utils.GetFullPath(mainFile))
		cmd.Stdout = nil
		cmd.Stderr = nil
		if err := cmd.Start(); err != nil {
			logger.Warn(fmt.Sprintf("Failed to start module [%s]", moduleName), err)
		} else {
			logger.Info(fmt.Sprintf("Started module [%s]", moduleName))
		}
	}()
}
