package utils

import (
	"os"
)

func DiscoverAllModules() ([]string, error) {
	dirs, err := os.ReadDir(GetFullPath("modules"))
	if err != nil {
		return nil, err
	}
	var modules []string
	for _, d := range dirs {
		if d.IsDir() {
			configPath := GetModuleConfigPath(d.Name())
			if _, err := os.Stat(configPath); err == nil {
				modules = append(modules, d.Name())
			}
		}
	}
	return modules, nil
}

func GetModuleRoute(defaultRoute string) string {
	if os.Getenv("GATEWAY_MODE") == "true" {
		return "/"
	}
	return defaultRoute
}
