package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

var globalConfig *AppConfig

func Init(path string) error {
	cfg, err := Load(path)
	if err != nil {
		return fmt.Errorf("❌ Read config error: %w", err)
	}
	globalConfig = cfg

	// Init viper config
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	vipererr := viper.ReadInConfig()
	if vipererr != nil {
		panic(fmt.Errorf("fatal error config file: %w", vipererr))
	}

	return nil
}

func Get() *AppConfig {
	if globalConfig == nil {
		panic("❌ Config not initialized. Did you forget to call config.Init(path)?")
	}
	return globalConfig
}

func Load(path string) (*AppConfig, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg AppConfig
	if err := yaml.Unmarshal(file, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
