package service

import (
	"context"
	"time"

	"github.com/khiemnd777/andy_api/modules/main/config"
	model "github.com/khiemnd777/andy_api/modules/main/features/__model"
	"github.com/khiemnd777/andy_api/modules/main/features/dashboard/case_daily_active_stats/repository"
	"github.com/khiemnd777/andy_api/shared/module"
	"github.com/khiemnd777/andy_api/shared/pubsub"
)

type CaseDailyActiveStatsService interface {
	UpsertOne(
		ctx context.Context,
		statDate time.Time,
		departmentID int,
	) error

	ActiveCases(
		ctx context.Context,
		departmentID *int,
		fromDate time.Time,
		toDate time.Time,
		previousFrom time.Time,
		previousTo time.Time,
	) (*model.ActiveCasesResult, error)
}

type caseDailyActiveStatsService struct {
	repo repository.CaseDailyActiveStatsRepository
	deps *module.ModuleDeps[config.ModuleConfig]
}

func NewCaseDailyActiveStatsService(
	repo repository.CaseDailyActiveStatsRepository,
	deps *module.ModuleDeps[config.ModuleConfig],
) CaseDailyActiveStatsService {
	svc := &caseDailyActiveStatsService{repo: repo, deps: deps}

	pubsub.SubscribeAsync("dashboard:daily:active:stats", func(payload *model.CaseDailyActiveStatsUpsert) error {
		ctx := context.Background()
		return svc.UpsertOne(ctx, payload.StatAt, payload.DepartmentID)
	})

	return svc
}

func (s *caseDailyActiveStatsService) UpsertOne(
	ctx context.Context,
	activeAt time.Time,
	departmentID int,
) error {
	return s.repo.UpsertOne(ctx, activeAt, departmentID)
}

func (s *caseDailyActiveStatsService) ActiveCases(
	ctx context.Context,
	departmentID *int, // nil = all departments
	fromDate time.Time,
	toDate time.Time,
	previousFrom time.Time,
	previousTo time.Time,
) (*model.ActiveCasesResult, error) {
	return s.repo.ActiveCases(ctx, departmentID, fromDate, toDate, previousFrom, previousTo)
}
