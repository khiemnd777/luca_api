package audit

import (
	"context"
	"net/http"
	"time"

	"github.com/khiemnd777/andy_api/shared/app"
	"github.com/khiemnd777/andy_api/shared/utils"
	"github.com/khiemnd777/andy_api/shared/worker"
)

/* How to use:
worker.Enqueue("audit", audit.LogRequest{
	UserID:   123,
	Action:   "delete",
	Module:   "user",
	TargetID: utils.Ptr(456),
})
*/

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
	// How to use:
	// c *fiber.Ctx <- Important note
	// err := audit.Send(c.UserContext(), token, audit.LogRequest{
	//   Action:   "update",
	//   Module:   "user",
	//   TargetID: &userID,
	//   ExtraData: map[string]any{
	//     "before": oldData,
	//     "after":  newData,
	//   },
	// })
	return app.GetHttpClient().CallRequest(
		ctx,
		http.MethodPost,
		"auditlog",
		"/api/audit/logs",
		"",
		utils.GetInternalToken(),
		log,
		nil,
		app.RetryOptions{
			MaxAttempts: 3,
			Delay:       200 * time.Millisecond,
		},
	)
}
