package service

import (
	"context"

	"github.com/khiemnd777/andy_api/modules/main/config"
	model "github.com/khiemnd777/andy_api/modules/main/features/__model"
	"github.com/khiemnd777/andy_api/modules/main/features/dashboard/activity_due_today/repository"
	"github.com/khiemnd777/andy_api/shared/module"
)

type DueTodayService interface {
	DueToday(
		ctx context.Context,
		deptID int,
	) ([]*model.DueTodayItem, error)
}

type dueTodayService struct {
	repo repository.DueTodayRepository
	deps *module.ModuleDeps[config.ModuleConfig]
}

func NewDueTodayService(
	repo repository.DueTodayRepository,
	deps *module.ModuleDeps[config.ModuleConfig],
) DueTodayService {
	return &dueTodayService{repo: repo, deps: deps}
}

func (s *dueTodayService) DueToday(
	ctx context.Context,
	deptID int,
) ([]*model.DueTodayItem, error) {
	return s.repo.DueToday(ctx, deptID)
}
