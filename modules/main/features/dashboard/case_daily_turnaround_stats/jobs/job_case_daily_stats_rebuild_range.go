package jobs

import (
	"context"
	"time"

	"github.com/khiemnd777/andy_api/modules/main/features/dashboard/case_daily_turnaround_stats/service"
	"github.com/khiemnd777/andy_api/shared/logger"
)

type CaseDailyStatsRebuildRangeJob struct {
	svc service.CaseDailyStatsService
}

func NewCaseDailyStatsRebuildRangeJob(svc service.CaseDailyStatsService) *CaseDailyStatsRebuildRangeJob {
	return &CaseDailyStatsRebuildRangeJob{svc: svc}
}

func (j CaseDailyStatsRebuildRangeJob) Name() string            { return "DashboardCaseDailyStatsRebuildRangeJob" }
func (j CaseDailyStatsRebuildRangeJob) DefaultSchedule() string { return "10 0 * * *" }
func (j CaseDailyStatsRebuildRangeJob) ConfigKey() string       { return "cron.dashboard_case_daily_stats" }

func (j CaseDailyStatsRebuildRangeJob) Run() error {
	logger.Debug("[DashboardCaseDailyStatsRebuildRangeJob] Dashboard case daily stats rebuilds range starting...")

	today := time.Now().UTC().Truncate(24 * time.Hour)

	if err := j.svc.RebuildRange(
		context.Background(),
		today.Add(-24*time.Hour),
		today.Add(24*time.Hour),
	); err != nil {
		logger.Error("[DashboardCaseDailyStatsRebuildRangeJob] Dashboard case daily stats rebuilds range failed", err)
		return err
	}

	logger.Debug("[DashboardCaseDailyStatsRebuildRangeJob] Done.")
	return nil
}
