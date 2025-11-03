### e.g.: worker audit log sending
1. Declare `modules/auditlog/worker/worker.go`

```golang
package audit

import (
	"context"
	"time"

	"github.com/khiemnd777/andy_api/shared/app"
	"github.com/khiemnd777/andy_api/shared/utils"
	"github.com/khiemnd777/andy_api/shared/worker"
)

type LogRequest struct {
	Action    string         `json:"action"`
	Module    string         `json:"module"`
	TargetID  *int           `json:"target_id,omitempty"`
	ExtraData map[string]any `json:"extra_data,omitempty"`
}

type auditSender struct{}

func (a auditSender) Send(ctx context.Context, log LogRequest) error {
	token := utils.GetAccessTokenFromContext(ctx)
	return Send(ctx, token, log)
}

func init() {
	q := worker.NewSenderWorker(1000, auditSender{})
	worker.RegisterEnqueuer("audit", q, q.Stop)
}

func Send(ctx context.Context, token string, log LogRequest) error {
	return app.GetHttpClient().CallPost(ctx,
		"auditlog",
		"/api/audit-logs",
		token,
		log,
		nil,
		app.RetryOptions{
			MaxAttempts: 3,
			Delay:       200 * time.Millisecond,
		},
	)
}

```

2. Usage

```golang
worker.Enqueue("audit", audit.LogRequest{
	UserID:   123,
	Action:   "delete",
	Module:   "user",
	TargetID: utils.Ptr(456),
})
```