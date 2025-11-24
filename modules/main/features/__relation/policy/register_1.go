package relation

import (
	"fmt"
	"sync"
)

var (
	mu1       sync.RWMutex
	registry1 = map[string]Config1{}
)

func Register1(key string, cfg Config1) {
	mu1.Lock()
	defer mu1.Unlock()
	if _, ok := registry1[key]; ok {
		panic("relation.Register: duplicate key " + key)
	}
	registry1[key] = cfg
}

func GetConfig1(key string) (Config1, error) {
	mu1.RLock()
	defer mu1.RUnlock()
	cfg, ok := registry1[key]
	if !ok {
		return Config1{}, fmt.Errorf("relation '%s' not registered", key)
	}
	return cfg, nil
}
