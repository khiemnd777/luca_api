// scripts/create_module/templates/config_config.go.tmpl
package config

import "github.com/khiemnd777/andy_api/shared/config"

type AuthConfig struct {
	Secret string `yaml:"secret"`
}

type ModuleConfig struct {
	Server   config.ServerConfig   `yaml:"server"`
	Database config.DatabaseConfig `yaml:"database"`
	Auth     AuthConfig            `yaml:"auth"`
}

func (c *ModuleConfig) GetServer() config.ServerConfig {
	return c.Server
}

func (c *ModuleConfig) GetDatabase() config.DatabaseConfig {
	return c.Database
}
