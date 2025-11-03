package cache

import (
	"time"

	"github.com/khiemnd777/andy_api/shared/redis"
)

func Get[T any](key string, ttl time.Duration, fallback func() (*T, error)) (*T, error) {
	return redis.GetOrSet("cache", key, ttl, fallback)
}

func GetList[T any](key string, ttl time.Duration, fallback func() ([]T, error)) ([]T, error) {
	wrappedFallback := func() (*[]T, error) {
		list, err := fallback()
		if err != nil {
			return nil, err
		}
		return &list, nil
	}

	ptr, err := redis.GetOrSet("cache", key, ttl, wrappedFallback)
	if err != nil {
		return nil, err
	}
	return *ptr, nil
}

type PaginatedCachedList[T any] struct {
	Items   []T  `json:"items"`
	HasMore bool `json:"hasMore"`
}

func GetListWithHasMore[T any](key string, ttl time.Duration, fallback func() ([]T, bool, error)) ([]T, bool, error) {
	wrapped := func() (*PaginatedCachedList[T], error) {
		items, hasMore, err := fallback()
		if err != nil {
			return nil, err
		}
		return &PaginatedCachedList[T]{Items: items, HasMore: hasMore}, nil
	}

	result, err := redis.GetOrSet("cache", key, ttl, wrapped)
	if err != nil {
		return nil, false, err
	}
	return result.Items, result.HasMore, nil
}
