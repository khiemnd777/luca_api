package relation

import (
	"fmt"
	"sync"
)

var (
	muRefSearch       sync.RWMutex
	registryRefSearch = map[string]ConfigSearch{}
)

func RegisterRefSearch(key string, cfg ConfigSearch) {
	muRefSearch.Lock()
	defer muRefSearch.Unlock()
	if _, ok := registryRefSearch[key]; ok {
		panic("relation.Register: duplicate key " + key)
	}
	registryRefSearch[key] = cfg
}

func GetConfigRefSearch(key string) (ConfigSearch, error) {
	muRefSearch.RLock()
	defer muRefSearch.RUnlock()
	cfg, ok := registryRefSearch[key]
	if !ok {
		return ConfigSearch{}, fmt.Errorf("relation '%s' not registered", key)
	}
	return cfg, nil
}
