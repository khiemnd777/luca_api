package config

import "github.com/khiemnd777/andy_api/shared/config"

type StorageConfig struct {
	PhotoPath string `yaml:"photo_path" mapstructure:"photo_path"`
}

type ModuleConfig struct {
	Server   config.ServerConfig   `yaml:"server"`
	Storage  StorageConfig         `yaml:"storage" mapstructure:"storage"`
	Database config.DatabaseConfig `yaml:"database"`
	Features struct {
		Enabled []string `mapstructure:"enabled"` // ["section","clinic"]
	} `mapstructure:"features"`
}

func (c *ModuleConfig) GetServer() config.ServerConfig {
	return c.Server
}

func (c *ModuleConfig) GetDatabase() config.DatabaseConfig {
	return c.Database
}
