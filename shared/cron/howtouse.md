### e.g.: job cleanup_token 
1. Config via `config.yaml`
```yaml
cron:
  cleanup_token: # ConfigKey
    enabled: true
    schedule: "@every 1h"
```
2. Declare `cleanup_token.go`
```golang
// 2. modules/auth/jobs/cleanup_token.go
package jobs

import (
	"github.com/khiemnd777/andy_api/shared/cron"
	"github.com/khiemnd777/andy_api/shared/logger"
)

type CleanupTokenJob struct{}

func (j CleanupTokenJob) Name() string           { return "CleanupExpiredToken" }
func (j CleanupTokenJob) DefaultSchedule() string { return "@every 1h" }
func (j CleanupTokenJob) ConfigKey() string      { return "cron.cleanup_token" }

func (j CleanupTokenJob) Run() error {
	logger.Info("[CleanupTokenJob] Cleaning expired tokens...")
	return nil
}

```

3. Call `cron.RegisterJob(CleanupTokenJob{})` in `main.go`