package service

import (
	"context"
	"time"

	"github.com/khiemnd777/andy_api/modules/main/config"
	model "github.com/khiemnd777/andy_api/modules/main/features/__model"
	"github.com/khiemnd777/andy_api/modules/main/features/dashboard/case_daily_sales_stats/repository"
	"github.com/khiemnd777/andy_api/shared/module"
)

type CaseDailySalesStatsService interface {
	UpsertOne(
		ctx context.Context,
		from time.Time,
		to time.Time,
	) error

	Summary(
		ctx context.Context,
		deptID int,
		from time.Time,
		to time.Time,
		prevFrom time.Time,
		prevTo time.Time,
	) (*model.SalesSummary, error)

	Daily(
		ctx context.Context,
		deptID int,
		from time.Time,
		to time.Time,
	) ([]*model.SalesDailyItem, error)
}

type caseDailySalesStatsService struct {
	repo repository.CaseDailySalesStatsRepository
	deps *module.ModuleDeps[config.ModuleConfig]
}

func NewCaseDailySalesStatsService(
	repo repository.CaseDailySalesStatsRepository,
	deps *module.ModuleDeps[config.ModuleConfig],
) CaseDailySalesStatsService {
	return &caseDailySalesStatsService{repo: repo, deps: deps}
}

func (s *caseDailySalesStatsService) UpsertOne(
	ctx context.Context,
	from time.Time,
	to time.Time,
) error {
	return s.repo.UpsertOne(ctx, from, to)
}

func (s *caseDailySalesStatsService) Summary(
	ctx context.Context,
	deptID int,
	from time.Time,
	to time.Time,
	prevFrom time.Time,
	prevTo time.Time,
) (*model.SalesSummary, error) {
	res, err := s.repo.Summary(ctx, deptID, from, to, prevFrom, prevTo)
	if err != nil {
		return nil, err
	}

	if res.PrevRevenue != 0 {
		percent := (res.TotalRevenue - res.PrevRevenue) / res.PrevRevenue * 100
		res.GrowthPercent = &percent
	}

	return res, nil
}

func (s *caseDailySalesStatsService) Daily(
	ctx context.Context,
	deptID int,
	from time.Time,
	to time.Time,
) ([]*model.SalesDailyItem, error) {
	return s.repo.Daily(ctx, deptID, from, to)
}
