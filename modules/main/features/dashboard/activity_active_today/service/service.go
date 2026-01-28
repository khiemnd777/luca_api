package service

import (
	"context"

	"github.com/khiemnd777/andy_api/modules/main/config"
	model "github.com/khiemnd777/andy_api/modules/main/features/__model"
	"github.com/khiemnd777/andy_api/modules/main/features/dashboard/activity_active_today/repository"
	"github.com/khiemnd777/andy_api/shared/module"
)

type ActiveTodayService interface {
	ActiveToday(
		ctx context.Context,
		deptID int,
	) ([]*model.ActiveTodayItem, error)
}

type activeTodayService struct {
	repo repository.ActiveTodayRepository
	deps *module.ModuleDeps[config.ModuleConfig]
}

func NewActiveTodayService(
	repo repository.ActiveTodayRepository,
	deps *module.ModuleDeps[config.ModuleConfig],
) ActiveTodayService {
	return &activeTodayService{repo: repo, deps: deps}
}

func (s *activeTodayService) ActiveToday(
	ctx context.Context,
	deptID int,
) ([]*model.ActiveTodayItem, error) {
	return s.repo.ActiveToday(ctx, deptID)
}
