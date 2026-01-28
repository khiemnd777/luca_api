package service

import (
	"context"
	"time"

	"github.com/khiemnd777/andy_api/modules/main/config"
	model "github.com/khiemnd777/andy_api/modules/main/features/__model"
	"github.com/khiemnd777/andy_api/modules/main/features/dashboard/case_daily_turnaround_stats/repository"
	"github.com/khiemnd777/andy_api/shared/module"
	"github.com/khiemnd777/andy_api/shared/pubsub"
)

type CaseDailyStatsService interface {
	UpsertOne(
		ctx context.Context,
		completedAt time.Time,
		departmentID int,
		turnaroundSec int64,
	) error

	RebuildRange(
		ctx context.Context,
		fromDate time.Time,
		toDate time.Time,
	) error

	AvgTurnaround(
		ctx context.Context,
		departmentID *int,
		fromDate time.Time,
		toDate time.Time,
		previousFrom time.Time,
		previousTo time.Time,
	) (*model.AvgTurnaroundResult, error)
}

type caseDailyStatsService struct {
	repo repository.CaseDailyStatsRepository
	deps *module.ModuleDeps[config.ModuleConfig]
}

func NewCaseDailyStatsService(
	repo repository.CaseDailyStatsRepository,
	deps *module.ModuleDeps[config.ModuleConfig],
) CaseDailyStatsService {
	svc := &caseDailyStatsService{repo: repo, deps: deps}

	pubsub.SubscribeAsync("dashboard:daily:stats", func(payload *model.CaseDailyStatsUpsert) error {
		ctx := context.Background()
		turnaroundsec := payload.CompletedAt.Sub(payload.ReceivedAt).Seconds()
		return svc.UpsertOne(ctx, payload.CompletedAt, payload.DepartmentID, int64(turnaroundsec))
	})

	return svc
}

func (s *caseDailyStatsService) UpsertOne(
	ctx context.Context,
	completedAt time.Time,
	departmentID int,
	turnaroundSec int64,
) error {
	return s.repo.UpsertOne(ctx, completedAt, departmentID, turnaroundSec)
}

func (s *caseDailyStatsService) RebuildRange(
	ctx context.Context,
	fromDate time.Time,
	toDate time.Time,
) error {
	return s.repo.RebuildRange(ctx, fromDate, toDate)
}

func (s *caseDailyStatsService) AvgTurnaround(
	ctx context.Context,
	departmentID *int,
	fromDate time.Time,
	toDate time.Time,
	previousFrom time.Time,
	previousTo time.Time,
) (*model.AvgTurnaroundResult, error) {
	return s.repo.AvgTurnaround(ctx, departmentID, fromDate, toDate, previousFrom, previousTo)
}
