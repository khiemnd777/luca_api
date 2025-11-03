package jobs

import (
	"context"

	"github.com/khiemnd777/andy_api/modules/auth_guest/service"
	"github.com/khiemnd777/andy_api/shared/logger"
)

type DeleteAllExpiredGuestsJob struct {
	svc *service.AuthGuestService
}

func NewDeleteAllExpiredGuestsJob(svc *service.AuthGuestService) *DeleteAllExpiredGuestsJob {
	return &DeleteAllExpiredGuestsJob{svc: svc}
}

func (j DeleteAllExpiredGuestsJob) Name() string            { return "DeleteAllExpiredGuests" }
func (j DeleteAllExpiredGuestsJob) DefaultSchedule() string { return "0 0 * * *" }
func (j DeleteAllExpiredGuestsJob) ConfigKey() string       { return "cron.delete_all_expired_guests" }

func (j DeleteAllExpiredGuestsJob) Run() error {
	logger.Debug("[DeleteAllExpiredGuestsJob] Delete all expired guests starting...")

	if _, err := j.svc.DeleteAllGuestsWithExpiredRefreshTokens(context.Background()); err != nil {
		logger.Error("[DeleteAllExpiredGuestsJob] Delete failed", err)
		return err
	}

	logger.Debug("[DeleteAllExpiredGuestsJob] Done.")
	return nil
}
