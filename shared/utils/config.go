package utils

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

func LoadConfig[T any](path string) (*T, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("❌ Failed to read config file %s: %w", path, err)
	}

	var config T
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("❌ Failed to unmarshal YAML config: %w", err)
	}

	return &config, nil
}
