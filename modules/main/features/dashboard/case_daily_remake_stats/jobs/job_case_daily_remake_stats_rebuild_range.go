package jobs

import (
	"context"
	"time"

	"github.com/khiemnd777/andy_api/modules/main/features/dashboard/case_daily_remake_stats/service"
	"github.com/khiemnd777/andy_api/shared/logger"
)

type CaseDailyRemakeStatsRebuildRangeJob struct {
	svc service.CaseDailyRemakeStatsService
}

func NewCaseDailyRemakeStatsRebuildRangeJob(svc service.CaseDailyRemakeStatsService) *CaseDailyRemakeStatsRebuildRangeJob {
	return &CaseDailyRemakeStatsRebuildRangeJob{svc: svc}
}

func (j CaseDailyRemakeStatsRebuildRangeJob) Name() string {
	return "DashboardCaseDailyRemakeStatsRebuildRangeJob"
}
func (j CaseDailyRemakeStatsRebuildRangeJob) DefaultSchedule() string { return "11 0 * * *" }
func (j CaseDailyRemakeStatsRebuildRangeJob) ConfigKey() string {
	return "cron.dashboard_case_daily_remake_stats"
}

func (j CaseDailyRemakeStatsRebuildRangeJob) Run() error {
	logger.Debug("[DashboardCaseDailyRemakeStatsRebuildRangeJob] Dashboard case daily remake stats rebuilds range starting...")

	today := time.Now().UTC().Truncate(24 * time.Hour)

	if err := j.svc.RebuildRange(
		context.Background(),
		today.Add(-24*time.Hour),
		today.Add(24*time.Hour),
	); err != nil {
		logger.Error("[DashboardCaseDailyRemakeStatsRebuildRangeJob] Dashboard case daily remake stats rebuilds range failed", err)
		return err
	}

	logger.Debug("[DashboardCaseDailyRemakeStatsRebuildRangeJob] Done.")
	return nil
}
