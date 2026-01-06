package jobs

import (
	"context"
	"fmt"

	"github.com/khiemnd777/andy_api/modules/main/features/order/service"
	"github.com/khiemnd777/andy_api/shared/logger"
)

type ClearExpiredOrderCodeJob struct {
	svc service.OrderCodeService
}

func NewClearExpiredOrderCodeJob(svc service.OrderCodeService) *ClearExpiredOrderCodeJob {
	return &ClearExpiredOrderCodeJob{svc: svc}
}

func (j ClearExpiredOrderCodeJob) Name() string            { return "ClearExpiredOrderCode" }
func (j ClearExpiredOrderCodeJob) DefaultSchedule() string { return "0 0 * * *" }
func (j ClearExpiredOrderCodeJob) ConfigKey() string       { return "cron.clear_expired_order_code" }

func (j ClearExpiredOrderCodeJob) Run() error {
	logger.Debug("[ClearExpiredOrderCodeJob] Clear expired order code starting...")

	if _, err := j.svc.CleanupExpiredReservations(context.Background()); err != nil {
		logger.Error(fmt.Sprintf("[ClearExpiredOrderCodeJob] Clear expired order code failed: %v", err))
		return err
	}

	logger.Debug("[ClearExpiredOrderCodeJob] Done.")
	return nil
}
