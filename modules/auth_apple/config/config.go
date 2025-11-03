package config

import "github.com/khiemnd777/andy_api/shared/config"

type AuthAppleConfig struct {
	FallbackEmailDomain string `yaml:"fallback_email_domain"`
}

type ModuleConfig struct {
	Server    config.ServerConfig   `yaml:"server"`
	Database  config.DatabaseConfig `yaml:"database"`
	AuthApple AuthAppleConfig       `yaml:"authapple"`
}

func (c *ModuleConfig) GetServer() config.ServerConfig {
	return c.Server
}

func (c *ModuleConfig) GetDatabase() config.DatabaseConfig {
	return c.Database
}
