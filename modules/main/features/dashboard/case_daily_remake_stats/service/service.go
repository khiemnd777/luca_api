package service

import (
	"context"
	"time"

	"github.com/khiemnd777/andy_api/modules/main/config"
	model "github.com/khiemnd777/andy_api/modules/main/features/__model"
	"github.com/khiemnd777/andy_api/modules/main/features/dashboard/case_daily_remake_stats/repository"
	"github.com/khiemnd777/andy_api/shared/module"
	"github.com/khiemnd777/andy_api/shared/pubsub"
)

type CaseDailyRemakeStatsService interface {
	UpsertOne(
		ctx context.Context,
		completedAt time.Time,
		departmentID int,
		isRemake bool,
	) error

	RebuildRange(
		ctx context.Context,
		fromDate time.Time,
		toDate time.Time,
	) error

	AvgRemakeRate(
		ctx context.Context,
		departmentID *int, // nil = all departments
		fromDate time.Time,
		toDate time.Time,
		previousFrom time.Time,
		previousTo time.Time,
	) (*model.AvgRemakeResult, error)
}

type caseDailyRemakeStatsService struct {
	repo repository.CaseDailyRemakeStatsRepository
	deps *module.ModuleDeps[config.ModuleConfig]
}

func NewCaseDailyRemakeStatsService(
	repo repository.CaseDailyRemakeStatsRepository,
	deps *module.ModuleDeps[config.ModuleConfig],
) CaseDailyRemakeStatsService {
	svc := &caseDailyRemakeStatsService{repo: repo, deps: deps}

	pubsub.SubscribeAsync("dashboard:daily:remake:stats", func(payload *model.CaseDailyRemakeStatsUpsert) error {
		ctx := context.Background()
		return svc.UpsertOne(ctx, payload.CompletedAt, payload.DepartmentID, payload.IsRemake)
	})

	return svc
}

func (s *caseDailyRemakeStatsService) UpsertOne(
	ctx context.Context,
	completedAt time.Time,
	departmentID int,
	isRemake bool,
) error {
	return s.repo.UpsertOne(ctx, completedAt, departmentID, isRemake)
}

func (s *caseDailyRemakeStatsService) RebuildRange(
	ctx context.Context,
	fromDate time.Time,
	toDate time.Time,
) error {
	return s.repo.RebuildRange(ctx, fromDate, toDate)
}

func (s *caseDailyRemakeStatsService) AvgRemakeRate(
	ctx context.Context,
	departmentID *int, // nil = all departments
	fromDate time.Time,
	toDate time.Time,
	previousFrom time.Time,
	previousTo time.Time,
) (*model.AvgRemakeResult, error) {
	return s.repo.AvgRemakeRate(ctx, departmentID, fromDate, toDate, previousFrom, previousTo)
}
