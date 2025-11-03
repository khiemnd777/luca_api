package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type ServerConfig struct {
	Server struct {
		Port int `yaml:"port"`
	} `yaml:"server"`
}

func main() {
	ports := []int{}

	// 📘 1. Read main config.yaml
	mainConfigPath := "config.yaml"
	if p := getPortFromConfig(mainConfigPath); p != 0 {
		ports = append(ports, p)
	}

	// 📘 2. Read modules/*/config.yaml
	files, err := filepath.Glob("modules/*/config.yaml")
	if err != nil {
		fmt.Println("❌ Failed to scan modules config:", err)
		os.Exit(1)
	}

	for _, path := range files {
		if p := getPortFromConfig(path); p != 0 {
			ports = append(ports, p)
		}
	}

	// 🧹 3. Kill each port
	for _, port := range ports {
		fmt.Printf("🔪 Killing processes on port %d...\n", port)
		cmd := exec.Command("sudo", "/usr/bin/fuser", "-k", fmt.Sprintf("%d/tcp", port))
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			fmt.Printf("⚠️ Port %d not in use or failed to kill: %v\n", port, err)
		} else {
			fmt.Printf("✅ Port %d cleaned up successfully.\n", port)
		}
	}
}

func getPortFromConfig(path string) int {
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("⚠️ Cannot read %s: %v\n", path, err)
		return 0
	}

	var conf ServerConfig
	if err := yaml.Unmarshal(data, &conf); err != nil {
		fmt.Printf("⚠️ Cannot parse YAML in %s: %v\n", path, err)
		return 0
	}

	return conf.Server.Port
}
