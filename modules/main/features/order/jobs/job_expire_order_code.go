package jobs

import (
	"context"
	"time"

	"github.com/khiemnd777/andy_api/modules/main/features/order/service"
	"github.com/khiemnd777/andy_api/shared/logger"
)

type ExpireOrderCodeJob struct {
	svc service.OrderCodeService
}

func NewExpireOrderCodeJob(svc service.OrderCodeService) *ExpireOrderCodeJob {
	return &ExpireOrderCodeJob{svc: svc}
}

func (j ExpireOrderCodeJob) Name() string            { return "ExpireOrderCode" }
func (j ExpireOrderCodeJob) DefaultSchedule() string { return "@every 17m" }
func (j ExpireOrderCodeJob) ConfigKey() string       { return "cron.expire_order_code" }

func (j ExpireOrderCodeJob) Run() error {
	logger.Debug("[ExpireOrderCodeJob] Expire order code starting...")

	if _, err := j.svc.ExpireReservations(context.Background(), time.Now()); err != nil {
		logger.Error("[ExpireOrderCodeJob] Expire order code failed", err)
		return err
	}

	logger.Debug("[ExpireOrderCodeJob] Done.")
	return nil
}
