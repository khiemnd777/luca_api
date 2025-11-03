package cache

import (
	"strings"

	"github.com/khiemnd777/andy_api/shared/logger"
	"github.com/khiemnd777/andy_api/shared/redis"
)

// UpdateAndInvalidate runs DB update and clears cache
func UpdateAndInvalidate(key string, updateFn func() error) error {
	/* How to use
	_ = cache.UpdateAndInvalidate("user:123", func() error {
		return repo.UpdateUser(ctx, "123", newData)
	})
	*/
	err := updateFn()
	if err != nil {
		return err
	}
	return redis.Del("cache", key)
}

// UpdateManyAndInvalidate runs updateFn and deletes multiple cache keys if update succeeds.
func UpdateManyAndInvalidate(keys []string, updateFn func() error) error {
	/* How to use
		_ = cache.UpdateManyAndInvalidate([]string{
		"user:123",
		"user:list:teamA",
	}, func() error {
		return repo.BatchUpdate(ctx, ids, data)
	})
	*/

	err := updateFn()
	if err != nil {
		return err
	}

	InvalidateKeys(keys...)

	return nil
}

func isPatternKey(key string) bool {
	return strings.Contains(key, "*")
}

// InvalidateKeys deletes multiple cache keys or patterns
func InvalidateKeys(keys ...string) {
	for _, key := range keys {
		if isPatternKey(key) {
			if err := redis.DelByPattern("cache", key); err != nil {
				logger.Warn("❌ Failed to invalidate pattern cache: " + key)
			} else {
				logger.Info("🧹 Pattern cache invalidated: " + key)
			}
		} else {
			if err := redis.Del("cache", key); err != nil {
				logger.Warn("❌ Failed to invalidate cache: " + key)
			} else {
				logger.Info("🧹 Cache invalidated: " + key)
			}
		}
	}
}
