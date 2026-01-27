package service

import (
	"context"
	"time"

	"github.com/khiemnd777/andy_api/modules/main/config"
	model "github.com/khiemnd777/andy_api/modules/main/features/__model"
	"github.com/khiemnd777/andy_api/modules/main/features/dashboard/case_daily_completed_stats/repository"
	"github.com/khiemnd777/andy_api/shared/module"
	"github.com/khiemnd777/andy_api/shared/pubsub"
)

type CaseDailyCompletedStatsService interface {
	UpsertOne(
		ctx context.Context,
		completedAt time.Time,
		departmentID int,
	) error

	RebuildRange(
		ctx context.Context,
		fromDate time.Time,
		toDate time.Time,
	) error

	CompletedCases(
		ctx context.Context,
		departmentID *int, // nil = all departments
		fromDate time.Time,
		toDate time.Time,
		previousFrom time.Time,
		previousTo time.Time,
	) (*model.ActiveCasesResult, error)
}

type caseDailyCompletedStatsService struct {
	repo repository.CaseDailyCompletedStatsRepository
	deps *module.ModuleDeps[config.ModuleConfig]
}

func NewCaseDailyCompletedStatsService(
	repo repository.CaseDailyCompletedStatsRepository,
	deps *module.ModuleDeps[config.ModuleConfig],
) CaseDailyCompletedStatsService {
	svc := &caseDailyCompletedStatsService{repo: repo, deps: deps}

	pubsub.SubscribeAsync("dashboard:daily:completed:stats", func(payload *model.CaseDailyCompletedStatsUpsert) error {
		ctx := context.Background()
		return svc.UpsertOne(ctx, payload.CompletedAt, payload.DepartmentID)
	})

	return svc
}

func (s *caseDailyCompletedStatsService) UpsertOne(
	ctx context.Context,
	completedAt time.Time,
	departmentID int,
) error {
	return s.repo.UpsertOne(ctx, completedAt, departmentID)
}

func (s *caseDailyCompletedStatsService) RebuildRange(
	ctx context.Context,
	fromDate time.Time,
	toDate time.Time,
) error {
	return s.repo.RebuildRange(ctx, fromDate, toDate)
}

func (s *caseDailyCompletedStatsService) CompletedCases(
	ctx context.Context,
	departmentID *int, // nil = all departments
	fromDate time.Time,
	toDate time.Time,
	previousFrom time.Time,
	previousTo time.Time,
) (*model.ActiveCasesResult, error) {
	return s.repo.CompletedCases(ctx, departmentID, fromDate, toDate, previousFrom, previousTo)
}
