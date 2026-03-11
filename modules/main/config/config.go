package config

import "github.com/khiemnd777/andy_api/shared/config"

type StorageConfig struct {
	PhotoPath string `yaml:"photo_path" mapstructure:"photo_path"`
}

type DeliveryQRConfig struct {
	SessionTTLMinutes   int    `yaml:"session_ttl_minutes" mapstructure:"session_ttl_minutes"`
	ClientBaseURL       string `yaml:"client_base_url" mapstructure:"client_base_url"`
	ProofImageMaxSizeMB int    `yaml:"proof_image_max_size_mb" mapstructure:"proof_image_max_size_mb"`
}

type ModuleConfig struct {
	Server     config.ServerConfig   `yaml:"server"`
	Storage    StorageConfig         `yaml:"storage" mapstructure:"storage"`
	Database   config.DatabaseConfig `yaml:"database"`
	DeliveryQR DeliveryQRConfig      `yaml:"delivery_qr" mapstructure:"delivery_qr"`
	Features   struct {
		Enabled []string `mapstructure:"enabled"` // ["section","clinic"]
	} `mapstructure:"features"`
}

func (c *ModuleConfig) GetServer() config.ServerConfig {
	return c.Server
}

func (c *ModuleConfig) GetDatabase() config.DatabaseConfig {
	return c.Database
}
