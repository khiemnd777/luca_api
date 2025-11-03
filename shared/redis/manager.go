package redis

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/khiemnd777/andy_api/shared/config"
	"github.com/khiemnd777/andy_api/shared/logger"

	"github.com/redis/go-redis/v9"
)

var (
	clients  = make(map[string]*redis.Client)
	ctx      = context.Background()
	initOnce sync.Once
)

func Init() {
	initOnce.Do(func() {
		for name, conf := range config.Get().Redis.Instances {
			addr := conf.Host + ":" + strconv.Itoa(conf.Port)
			opts := &redis.Options{
				Addr:     addr,
				Password: conf.Password,
				DB:       conf.DB,
			}
			client := redis.NewClient(opts)
			if err := client.Ping(context.Background()).Err(); err != nil {
				log.Printf("⚠️ Redis [%s] not connected: %v", name, err)
			} else {
				clients[name] = client
				log.Printf("✅ Redis [%s] connected: %s", name, addr)
			}
		}
	})
}

// Get returns redis client by instance name (e.g. "cache", "session")
func GetInstance(name string) *redis.Client {
	return clients[name]
}

func Get(name, key string) (string, error) {
	rdb := GetInstance(name)
	val, err := rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		logger.Info("🔍 Redis GET: key not found: " + key)
		return "", nil
	} else if err != nil {
		logger.Warn("❌ Redis GET error: "+key, err)
		return "", err
	}
	return val, nil
}

func GetInt(name, key string) (int, error) {
	val, err := Get(name, key)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(val)
}

func Set(name, key string, value interface{}, ttl time.Duration) error {
	rdb := GetInstance(name)
	err := rdb.Set(ctx, key, value, ttl).Err()
	if err != nil {
		logger.Warn("❌ Redis SET error: "+key, err)
	}
	return err
}

func Incr(name, key string) (int64, error) {
	rdb := GetInstance(name)
	incr := rdb.Incr(ctx, key)
	val, err := incr.Result()
	if err != nil {
		logger.Warn("❌ Redis SET error: "+key, err)
	}
	return val, err
}

func Del(name, key string) error {
	rdb := GetInstance(name)
	err := rdb.Del(ctx, key).Err()
	if err != nil {
		logger.Warn("❌ Redis DEL error: "+key, err)
	}
	return err
}

func DelByPattern(name string, pattern string) error {
	rdb := GetInstance(name)

	var cursor uint64
	ctx := context.Background()

	for {
		keys, nextCursor, err := rdb.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return fmt.Errorf("error scanning keys with pattern %s: %w", pattern, err)
		}

		if len(keys) > 0 {
			if _, err := rdb.Del(ctx, keys...).Result(); err != nil {
				return fmt.Errorf("error deleting keys: %w", err)
			}
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	return nil
}

func Exists(name, key string) (bool, error) {
	rdb := GetInstance(name)
	count, err := rdb.Exists(ctx, key).Result()
	if err != nil {
		logger.Warn("❌ Redis EXISTS error: "+key, err)
		return false, err
	}
	return count > 0, nil
}

func ScanKeys(name, pattern string) ([]string, error) {
	rdb := GetInstance(name)
	var cursor uint64
	var keys []string
	var err error
	for {
		var k []string
		k, cursor, err = rdb.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return nil, err
		}
		keys = append(keys, k...)
		if cursor == 0 {
			break
		}
	}
	return keys, nil
}

func TTL(name, key string) (time.Duration, error) {
	rdb := GetInstance(name)
	ttl, err := rdb.TTL(ctx, key).Result()
	if err != nil {
		logger.Warn("❌ Redis TTL error: "+key, err)
		return 0, err
	}
	return ttl, nil
}

func InitFromConfig(cfg config.RedisConfig) {
	mu := sync.Mutex{}
	mu.Lock()
	defer mu.Unlock()

	for name, rc := range cfg.Instances {
		addr := fmt.Sprintf("%s:%d", rc.Host, rc.Port)
		client := redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: rc.Password,
			DB:       rc.DB,
		})

		if err := client.Ping(context.Background()).Err(); err != nil {
			logger.Warn(fmt.Sprintf("⚠️ Redis [%s] not connected: %v", name, err))
		} else {
			clients[name] = client
			logger.Info(fmt.Sprintf("✅ Redis [%s] connected: %s", name, addr))
		}
	}
}
