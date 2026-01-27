package order

import (
	"github.com/gofiber/fiber/v2"

	"github.com/khiemnd777/andy_api/modules/main/config"
	dailyactivehlr "github.com/khiemnd777/andy_api/modules/main/features/dashboard/case_daily_active_stats/handler"
	dailyactiverepo "github.com/khiemnd777/andy_api/modules/main/features/dashboard/case_daily_active_stats/repository"
	dailyactivesvc "github.com/khiemnd777/andy_api/modules/main/features/dashboard/case_daily_active_stats/service"
	dailycompletedhlr "github.com/khiemnd777/andy_api/modules/main/features/dashboard/case_daily_completed_stats/handler"
	dailycompletedrepo "github.com/khiemnd777/andy_api/modules/main/features/dashboard/case_daily_completed_stats/repository"
	dailycompletedsvc "github.com/khiemnd777/andy_api/modules/main/features/dashboard/case_daily_completed_stats/service"
	dailyremakehlr "github.com/khiemnd777/andy_api/modules/main/features/dashboard/case_daily_remake_stats/handler"
	dailyremakerepo "github.com/khiemnd777/andy_api/modules/main/features/dashboard/case_daily_remake_stats/repository"
	dailyremakesvc "github.com/khiemnd777/andy_api/modules/main/features/dashboard/case_daily_remake_stats/service"
	turnaroundhlr "github.com/khiemnd777/andy_api/modules/main/features/dashboard/case_daily_turnaround_stats/handler"
	turnaroundrepo "github.com/khiemnd777/andy_api/modules/main/features/dashboard/case_daily_turnaround_stats/repository"
	turnaroundsvc "github.com/khiemnd777/andy_api/modules/main/features/dashboard/case_daily_turnaround_stats/service"
	casestatuseshlr "github.com/khiemnd777/andy_api/modules/main/features/dashboard/case_statuses/handler"
	casestatusesrepo "github.com/khiemnd777/andy_api/modules/main/features/dashboard/case_statuses/repository"
	casestatusessvc "github.com/khiemnd777/andy_api/modules/main/features/dashboard/case_statuses/service"
	duetodayhlr "github.com/khiemnd777/andy_api/modules/main/features/dashboard/due_today/handler"
	duetodayrepo "github.com/khiemnd777/andy_api/modules/main/features/dashboard/due_today/repository"
	duetodaysvc "github.com/khiemnd777/andy_api/modules/main/features/dashboard/due_today/service"
	"github.com/khiemnd777/andy_api/modules/main/registry"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/metadata/customfields"
	"github.com/khiemnd777/andy_api/shared/module"
)

type feature struct{}

func (feature) ID() string    { return "dashboard" }
func (feature) Priority() int { return 69 }

func (feature) Register(router fiber.Router, deps *module.ModuleDeps[config.ModuleConfig], cfMgr *customfields.Manager) error {
	entClient := deps.Ent.(*generated.Client)
	// Case Daily Stats
	caseDailyStatsRepo := turnaroundrepo.NewCaseDailyStatsRepository(entClient, deps.DB, deps)
	caseDailyStatsSvc := turnaroundsvc.NewCaseDailyStatsService(caseDailyStatsRepo, deps)
	caseDailyStatsHandler := turnaroundhlr.NewCaseDailyStatsHandler(caseDailyStatsSvc, deps)
	caseDailyStatsHandler.RegisterRoutes(router)
	// cron.RegisterJob(jobs.NewCaseDailyStatsRebuildRangeJob(caseDailyStatsSvc))

	// Case Daily Remake Stats
	caseDailyRemakeStatsRepo := dailyremakerepo.NewCaseDailyRemakeStatsRepository(entClient, deps.DB, deps)
	caseDailyRemakeStatsSvc := dailyremakesvc.NewCaseDailyRemakeStatsService(caseDailyRemakeStatsRepo, deps)
	caseDailyRemakeStatsHandler := dailyremakehlr.NewCaseDailyRemakeStatsHandler(caseDailyRemakeStatsSvc, deps)
	caseDailyRemakeStatsHandler.RegisterRoutes(router)
	// cron.RegisterJob(dailyremakejobs.NewCaseDailyRemakeStatsRebuildRangeJob(caseDailyRemakeStatsSvc))

	// Case Daily Completed Stats
	caseDailyCompletedStatsRepo := dailycompletedrepo.NewCaseDailyCompletedStatsRepository(entClient, deps.DB, deps)
	caseDailyCompletedStatsSvc := dailycompletedsvc.NewCaseDailyCompletedStatsService(caseDailyCompletedStatsRepo, deps)
	caseDailyCompletedStatsHandler := dailycompletedhlr.NewCaseDailyCompletedStatsHandler(caseDailyCompletedStatsSvc, deps)
	caseDailyCompletedStatsHandler.RegisterRoutes(router)
	// cron.RegisterJob(dailycompletedjobs.NewCaseDailyCompletedStatsRebuildRangeJob(caseDailyCompletedStatsSvc))

	// Case Daily Active Stats
	caseDailyActiveStatsRepo := dailyactiverepo.NewCaseDailyActiveStatsRepository(entClient, deps.DB, deps)
	caseDailyActiveStatsSvc := dailyactivesvc.NewCaseDailyActiveStatsService(caseDailyActiveStatsRepo, deps)
	caseDailyActiveStatsHandler := dailyactivehlr.NewCaseDailyActiveStatsHandler(caseDailyActiveStatsSvc, deps)
	caseDailyActiveStatsHandler.RegisterRoutes(router)

	// Due Today
	duetodayRepo := duetodayrepo.NewDueTodayRepository(entClient, deps.DB, deps)
	duetodaySvc := duetodaysvc.NewDueTodayService(duetodayRepo, deps)
	duetodayHandler := duetodayhlr.NewDueTodayHandler(duetodaySvc, deps)
	duetodayHandler.RegisterRoutes(router)

	// Case Statuses
	caseStatusesRepo := casestatusesrepo.NewCaseStatusesRepository(entClient, deps.DB, deps)
	caseStatusesSvc := casestatusessvc.NewCaseStatusesService(caseStatusesRepo, deps)
	caseStatusesHandler := casestatuseshlr.NewCaseStatusesHandler(caseStatusesSvc, deps)
	caseStatusesHandler.RegisterRoutes(router)

	return nil
}

func init() { registry.Register(feature{}) }
