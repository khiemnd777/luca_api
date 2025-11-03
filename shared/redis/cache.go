package redis

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/khiemnd777/andy_api/shared/logger"
)

// GetOrSet generic cache-aside for any type T
func GetOrSet[T any](name, key string, ttl time.Duration, fallback func() (*T, error)) (*T, error) {
	val, err := Get(name, key)
	if err != nil {
		return nil, err
	}
	if val != "" {
		logger.Info("⚡ Cache HIT: " + key)
		var result T
		if err := json.Unmarshal([]byte(val), &result); err != nil {
			logger.Warn("❌ Failed to unmarshal cache: "+key, err)
			return nil, err
		}
		return &result, nil
	}

	logger.Info("🌀 Cache MISS: " + key + " → fallback")

	result, err := fallback()
	if err != nil {
		return nil, err
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		logger.Warn("❌ Failed to marshal cache: "+key, err)
		return nil, err
	}

	if err := Set(name, key, string(jsonData), ttl); err != nil {
		logger.Warn(fmt.Sprintf("⚠️ Redis[%s] failed to cache %s: %v", name, key, err))
		return nil, err
	}

	return result, nil
}
