package relation

import (
	"fmt"
	"sync"
)

var (
	mu1N       sync.RWMutex
	registry1N = map[string]Config1N{}
)

func Register1N(key string, cfg Config1N) {
	mu1N.Lock()
	defer mu1N.Unlock()
	if _, ok := registry1N[key]; ok {
		panic("relation.Register: duplicate key " + key)
	}
	registry1N[key] = cfg
}

func GetConfig1N(key string) (Config1N, error) {
	mu1N.RLock()
	defer mu1N.RUnlock()
	cfg, ok := registry1N[key]
	if !ok {
		return Config1N{}, fmt.Errorf("relation '%s' not registered", key)
	}
	return cfg, nil
}
