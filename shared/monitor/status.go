package monitor

import (
	"context"
	"encoding/json"
	"time"

	"github.com/khiemnd777/andy_api/shared/logger"
	"github.com/khiemnd777/andy_api/shared/redis"
)

func SetModuleStatus(name string, port int, status string) {
	rdb := redis.GetInstance("status")
	if rdb == nil {
		logger.Warn("⚠️ Redis instance 'status' not available")
		return
	}

	ctx := context.Background()
	data := map[string]interface{}{
		"name":      name,
		"port":      port,
		"status":    status,
		"updatedAt": time.Now(),
	}

	jsonData, _ := json.Marshal(data)
	if err := rdb.Set(ctx, "module_status:"+name, jsonData, 0).Err(); err != nil {
		logger.Warn("❌ Failed to write module status to Redis", err)
	}
}
