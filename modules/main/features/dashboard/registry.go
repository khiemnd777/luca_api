package order

import (
	"github.com/gofiber/fiber/v2"

	"github.com/khiemnd777/andy_api/modules/main/config"
	dailyremakehlr "github.com/khiemnd777/andy_api/modules/main/features/dashboard/case_daily_remake_stats/handler"
	dailyremakejobs "github.com/khiemnd777/andy_api/modules/main/features/dashboard/case_daily_remake_stats/jobs"
	dailyremakerepo "github.com/khiemnd777/andy_api/modules/main/features/dashboard/case_daily_remake_stats/repository"
	dailyremakesvc "github.com/khiemnd777/andy_api/modules/main/features/dashboard/case_daily_remake_stats/service"
	"github.com/khiemnd777/andy_api/modules/main/features/dashboard/case_daily_stats/handler"
	"github.com/khiemnd777/andy_api/modules/main/features/dashboard/case_daily_stats/jobs"
	"github.com/khiemnd777/andy_api/modules/main/features/dashboard/case_daily_stats/repository"
	"github.com/khiemnd777/andy_api/modules/main/features/dashboard/case_daily_stats/service"
	"github.com/khiemnd777/andy_api/modules/main/registry"
	"github.com/khiemnd777/andy_api/shared/cron"
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
	caseDailyStatsRepo := repository.NewCaseDailyStatsRepository(entClient, deps.DB, deps)
	caseDailyStatsSvc := service.NewCaseDailyStatsService(caseDailyStatsRepo, deps)
	caseDailyStatsHandler := handler.NewCaseDailyStatsHandler(caseDailyStatsSvc, deps)
	caseDailyStatsHandler.RegisterRoutes(router)
	cron.RegisterJob(jobs.NewCaseDailyStatsRebuildRangeJob(caseDailyStatsSvc))

	// Case Daily Remake Stats
	caseDailyRemakeStatsRepo := dailyremakerepo.NewCaseDailyRemakeStatsRepository(entClient, deps.DB, deps)
	caseDailyRemakeStatsSvc := dailyremakesvc.NewCaseDailyRemakeStatsService(caseDailyRemakeStatsRepo, deps)
	caseDailyRemakeStatsHandler := dailyremakehlr.NewCaseDailyRemakeStatsHandler(caseDailyRemakeStatsSvc, deps)
	caseDailyRemakeStatsHandler.RegisterRoutes(router)
	cron.RegisterJob(dailyremakejobs.NewCaseDailyRemakeStatsRebuildRangeJob(caseDailyRemakeStatsSvc))

	return nil
}

func init() { registry.Register(feature{}) }
