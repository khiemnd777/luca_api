package jobs

import (
	"context"

	"github.com/khiemnd777/andy_api/modules/main/features/dashboard/case_daily_remake_stats/service"
	"github.com/khiemnd777/andy_api/shared/logger"
	"github.com/khiemnd777/andy_api/shared/utils"
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

	from, to := utils.DayRange(-1, 1)

	if err := j.svc.RebuildRange(
		context.Background(),
		from,
		to,
	); err != nil {
		logger.Error("[DashboardCaseDailyRemakeStatsRebuildRangeJob] Dashboard case daily remake stats rebuilds range failed", err)
		return err
	}

	logger.Debug("[DashboardCaseDailyRemakeStatsRebuildRangeJob] Done.")
	return nil
}
