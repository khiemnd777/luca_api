package service

import (
	"context"
	"time"

	"github.com/khiemnd777/andy_api/modules/main/config"
	model "github.com/khiemnd777/andy_api/modules/main/features/__model"
	"github.com/khiemnd777/andy_api/modules/main/features/dashboard/case_daily_sales_stats/repository"
	"github.com/khiemnd777/andy_api/shared/logger"
	"github.com/khiemnd777/andy_api/shared/module"
	"github.com/khiemnd777/andy_api/shared/pubsub"
)

type CaseDailySalesStatsService interface {
	UpsertOne(
		ctx context.Context,
		deptID int,
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

	GetReport(
		ctx context.Context,
		deptID int,
		r model.Range,
	) (*model.SalesReportResponse, error)
}

type caseDailySalesStatsService struct {
	repo repository.CaseDailySalesStatsRepository
	deps *module.ModuleDeps[config.ModuleConfig]
}

func NewCaseDailySalesStatsService(
	repo repository.CaseDailySalesStatsRepository,
	deps *module.ModuleDeps[config.ModuleConfig],
) CaseDailySalesStatsService {
	svc := &caseDailySalesStatsService{repo: repo, deps: deps}

	pubsub.SubscribeAsync("dashboard:daily:sales", func(payload *model.SalesDailyUpsert) error {
		if payload == nil {
			logger.Warn("sales_daily_upsert: payload is nil")
			return nil // swallow, tránh crash subscriber
		}
		ctx := context.Background()
		statAt := payload.StatAt
		from := time.Date(
			statAt.Year(),
			statAt.Month(),
			statAt.Day(),
			0, 0, 0, 0,
			statAt.Location(),
		)

		to := from.AddDate(0, 0, 1)

		logger.Debug(
			"sales_daily_upsert",
			"department_id", payload.DepartmentID,
			"from", from.Format("2006-01-02"),
			"to", to.Format("2006-01-02"),
		)

		if err := svc.UpsertOne(ctx, payload.DepartmentID, from, to); err != nil {
			logger.Error(
				"sales_daily_upsert_failed",
				"department_id", payload.DepartmentID,
				"stat_date", from.Format("2006-01-02"),
				"error", err,
			)
			return err
		}
		return nil
	})

	return svc
}

func (s *caseDailySalesStatsService) UpsertOne(
	ctx context.Context,
	deptID int,
	from time.Time,
	to time.Time,
) error {
	return s.repo.UpsertOne(ctx, deptID, from, to)
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

type SalesReportResponse struct {
	KPIs *model.SalesSummary     `json:"kpis,omitempty"`
	Line []*model.SalesDailyItem `json:"line,omitempty"`
}

func (s *caseDailySalesStatsService) GetReport(
	ctx context.Context,
	deptID int,
	r model.Range,
) (*model.SalesReportResponse, error) {

	now := time.Now()
	from, to, prevFrom, prevTo := resolveRange(r, now)

	summary, err := s.Summary(ctx, deptID, from, to, prevFrom, prevTo)
	if err != nil {
		return nil, err
	}

	daily, err := s.Daily(ctx, deptID, from, to)
	if err != nil {
		return nil, err
	}

	return &model.SalesReportResponse{
		KPIs: summary,
		Line: daily,
	}, nil
}

func resolveRange(r model.Range, now time.Time) (
	from, to time.Time,
	prevFrom, prevTo time.Time,
) {
	end := now
	var days int

	switch r {
	case model.RangeToday:
		days = 1
	case model.Range7d:
		days = 7
	case model.Range30d:
		days = 30
	}

	from = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).
		AddDate(0, 0, -(days - 1))
	to = end

	prevTo = from.AddDate(0, 0, -1)
	prevFrom = prevTo.AddDate(0, 0, -(days - 1))

	return
}
