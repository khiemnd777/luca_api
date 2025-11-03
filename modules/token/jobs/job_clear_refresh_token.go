package jobs

import (
	"context"

	"github.com/khiemnd777/andy_api/modules/token/service"
	"github.com/khiemnd777/andy_api/shared/logger"
)

type ClearResfreshTokenJob struct {
	svc *service.TokenService
}

func NewClearResfreshTokenJob(svc *service.TokenService) *ClearResfreshTokenJob {
	return &ClearResfreshTokenJob{svc: svc}
}

func (j ClearResfreshTokenJob) Name() string            { return "ClearResfreshToken" }
func (j ClearResfreshTokenJob) DefaultSchedule() string { return "@every 1h" }
func (j ClearResfreshTokenJob) ConfigKey() string       { return "cron.clear_refresh_token" }

func (j ClearResfreshTokenJob) Run() error {
	logger.Debug("[ClearResfreshTokenJob] Clear refresh token starting...")

	if err := j.svc.CleanupExpiredRefreshTokens(context.Background()); err != nil {
		logger.Error("[ClearResfreshTokenJob] Clear refresh token failed", err)
		return err
	}

	logger.Debug("[ClearResfreshTokenJob] Done.")
	return nil
}
