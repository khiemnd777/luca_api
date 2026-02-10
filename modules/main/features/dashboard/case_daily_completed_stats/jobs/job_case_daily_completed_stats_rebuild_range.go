package jobs

import (
	"context"

	"github.com/khiemnd777/andy_api/modules/main/features/dashboard/case_daily_completed_stats/service"
	"github.com/khiemnd777/andy_api/shared/logger"
	"github.com/khiemnd777/andy_api/shared/utils"
)

type CaseDailyCompletedStatsRebuildRangeJob struct {
	svc service.CaseDailyCompletedStatsService
}

func NewCaseDailyCompletedStatsRebuildRangeJob(svc service.CaseDailyCompletedStatsService) *CaseDailyCompletedStatsRebuildRangeJob {
	return &CaseDailyCompletedStatsRebuildRangeJob{svc: svc}
}

func (j CaseDailyCompletedStatsRebuildRangeJob) Name() string {
	return "DashboardCaseDailyCompletedStatsRebuildRangeJob"
}
func (j CaseDailyCompletedStatsRebuildRangeJob) DefaultSchedule() string { return "12 0 * * *" }
func (j CaseDailyCompletedStatsRebuildRangeJob) ConfigKey() string {
	return "cron.dashboard_case_daily_completed_stats"
}

func (j CaseDailyCompletedStatsRebuildRangeJob) Run() error {
	logger.Debug("[DashboardCaseDailyCompletedStatsRebuildRangeJob] Dashboard case daily completed stats rebuilds range starting...")

	from, to := utils.DayRange(-1, 1)

	if err := j.svc.RebuildRange(
		context.Background(),
		from,
		to,
	); err != nil {
		logger.Error("[DashboardCaseDailyCompletedStatsRebuildRangeJob] Dashboard case daily completed stats rebuilds range failed", err)
		return err
	}

	logger.Debug("[DashboardCaseDailyCompletedStatsRebuildRangeJob] Done.")
	return nil
}
