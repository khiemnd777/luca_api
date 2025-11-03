package utils

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// CheckPortOpen returns true if host:port is reachable via TCP
func CheckPortOpen(host string, port int) bool {
	addr := net.JoinHostPort(host, fmt.Sprintf("%d", port))
	conn, err := net.DialTimeout("tcp", addr, 1*time.Second)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

// DetectPIDFromPort attempts to find PID listening on given port (cross-platform)
func DetectPIDFromPort(port int) (int, error) {
	switch runtime.GOOS {
	case "windows":
		return detectPIDWindows(port)
	case "linux", "darwin":
		return detectPIDUnix(port)
	default:
		return 0, fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
}

func detectPIDUnix(port int) (int, error) {
	cmd := exec.Command("lsof", "-i", fmt.Sprintf(":%d", port), "-t")
	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("lsof failed: %w", err)
	}

	pidStr := strings.TrimSpace(string(output))
	if pidStr == "" {
		return 0, errors.New("no PID found")
	}
	return strconv.Atoi(pidStr)
}

func detectPIDWindows(port int) (int, error) {
	cmd := exec.Command("netstat", "-ano")
	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("netstat failed: %w", err)
	}

	lines := bytes.Split(output, []byte("\n"))
	target := fmt.Sprintf(":%d", port)

	for _, line := range lines {
		str := string(line)
		if strings.Contains(str, target) && strings.Contains(str, "LISTENING") {
			fields := strings.Fields(str)
			if len(fields) >= 5 {
				pidStr := fields[len(fields)-1]
				return strconv.Atoi(pidStr)
			}
		}
	}
	return 0, fmt.Errorf("no matching PID found for port %d", port)
}
