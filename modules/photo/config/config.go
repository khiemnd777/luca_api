package config

import "github.com/khiemnd777/andy_api/shared/config"

type StorageConfig struct {
	PhotoPath string `yaml:"photo_path"`
}

type ModuleConfig struct {
	Server   config.ServerConfig   `yaml:"server"`
	Storage  StorageConfig         `yaml:"storage"`
	Database config.DatabaseConfig `yaml:"database"`
}

func (c *ModuleConfig) GetServer() config.ServerConfig {
	return c.Server
}

func (c *ModuleConfig) GetDatabase() config.DatabaseConfig {
	return c.Database
}
