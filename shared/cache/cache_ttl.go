package cache

import (
	"time"

	"github.com/khiemnd777/andy_api/shared/config"
)

var (
	TTLShort  time.Duration
	TTLMedium time.Duration
	TTLLong   time.Duration
	TTLStatic time.Duration
)

func InitTTLConstants() {
	ttlCfg := config.Get().Cache.TTL
	TTLShort = ttlCfg.Short
	TTLMedium = ttlCfg.Medium
	TTLLong = ttlCfg.Long
	TTLStatic = ttlCfg.Static
}
