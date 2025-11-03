// scripts/create_module/templates/config_config.go.tmpl
package config

import "github.com/khiemnd777/andy_api/shared/config"

type AuthGoogleConfig struct {
	RedirectUri     string `yaml:"redirecturi"`
	ClientID        string `yaml:"clientid"`
	ClientSecret    string `yaml:"clientsecret"`
	AuthURL         string `yaml:"authurl"`
	TokenURL        string `yaml:"tokenurl"`
	AppRedirectBase string `yaml:"appredirectbase"`
}

type ModuleConfig struct {
	Server     config.ServerConfig   `yaml:"server"`
	Database   config.DatabaseConfig `yaml:"database"`
	AuthGoogle AuthGoogleConfig      `yaml:"authgoogle"`
}

func (c *ModuleConfig) GetServer() config.ServerConfig {
	return c.Server
}

func (c *ModuleConfig) GetDatabase() config.DatabaseConfig {
	return c.Database
}
