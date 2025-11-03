package jobs

import (
	"context"

	"github.com/khiemnd777/andy_api/modules/photo/service"
	"github.com/khiemnd777/andy_api/shared/logger"
)

type ClearDeletedPhotoJob struct {
	svc *service.PhotoService
}

func NewClearDeletedPhotoJob(svc *service.PhotoService) *ClearDeletedPhotoJob {
	return &ClearDeletedPhotoJob{svc: svc}
}

func (j ClearDeletedPhotoJob) Name() string            { return "ClearDeletedPhoto" }
func (j ClearDeletedPhotoJob) DefaultSchedule() string { return "0 0 * * *" } // 0h hàng ngày
func (j ClearDeletedPhotoJob) ConfigKey() string       { return "cron.clear_deleted_photo" }

func (j ClearDeletedPhotoJob) Run() error {
	logger.Debug("[ClearDeletedPhotoJob] Clearing deleted photos...")

	if err := j.svc.CleanDeletedPhotoFiles(context.Background()); err != nil {
		logger.Error("[ClearDeletedPhotoJob] Clearning failed.", err)
		return err
	}

	logger.Debug("[ClearDeletedPhotoJob] Done.")
	return nil
}
