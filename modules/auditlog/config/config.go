// scripts/create_module/templates/config_config.go.tmpl
package config

import "github.com/khiemnd777/andy_api/shared/config"

type ModuleConfig struct {
	Server         config.ServerConfig   `yaml:"server"`
	Database       config.DatabaseConfig `yaml:"database"`
	SharedDatabase config.DatabaseConfig `yaml:"shared_database"`
}

func (c *ModuleConfig) GetServer() config.ServerConfig {
	return c.Server
}

func (c *ModuleConfig) GetDatabase() config.DatabaseConfig {
	return c.Database
}

func (c *ModuleConfig) GetSharedDatabase() config.DatabaseConfig {
	return c.SharedDatabase
}
