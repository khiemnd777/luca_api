package registry

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/khiemnd777/andy_api/shared/runtime"
	"gopkg.in/yaml.v3"
)

type rawConfig struct {
	Server struct {
		Host  string `yaml:"host"`
		Port  int    `yaml:"port"`
		Route string `yaml:"route"`
	} `yaml:"server"`
	External bool `yaml:"external"`
}

type ModuleConfig struct {
	Name     string
	Host     string
	Port     int
	Route    string
	External bool
}

func LoadAllModules(dir string) ([]ModuleConfig, error) {
	// 1) Đọc registry dynamic trước
	reg, _ := runtime.LoadRegistry() // map[name]→RunningModule

	var modules []ModuleConfig

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}

		name := e.Name()
		configPath := filepath.Join(dir, name, "config.yaml")
		data, err := os.ReadFile(configPath)
		if err != nil {
			continue // skip module without config
		}

		var raw rawConfig
		if err := yaml.Unmarshal(data, &raw); err != nil {
			continue
		}

		// --- Lấy host/port ---
		host := raw.Server.Host
		port := raw.Server.Port

		if rm, ok := reg[name]; ok {
			// runtime entry → override host & port
			host = rm.Host
			port = rm.Port
		}

		// Skip module nếu port chưa xác định
		if port == 0 {
			continue
		}

		modules = append(modules, ModuleConfig{
			Name:     name,
			Port:     port,
			Host:     fmt.Sprintf("http://%s:%d", host, port),
			Route:    raw.Server.Route,
			External: raw.External,
		})
	}

	return modules, nil
}
