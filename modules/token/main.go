package main

import (
	"database/sql"

	entsql "entgo.io/ent/dialect/sql"

	"github.com/gofiber/fiber/v2"
	"github.com/khiemnd777/andy_api/modules/token/config"
	"github.com/khiemnd777/andy_api/modules/token/handler"
	"github.com/khiemnd777/andy_api/modules/token/jobs"
	"github.com/khiemnd777/andy_api/modules/token/repository"
	"github.com/khiemnd777/andy_api/modules/token/service"
	"github.com/khiemnd777/andy_api/shared/cron"
	"github.com/khiemnd777/andy_api/shared/db/ent"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/middleware"
	"github.com/khiemnd777/andy_api/shared/module"
	"github.com/khiemnd777/andy_api/shared/utils"
)

func main() {
	module.StartModule(module.ModuleOptions[config.ModuleConfig]{
		ConfigPath: utils.GetModuleConfigPath("token"),
		ModuleName: "token",
		InitEntClient: func(provider string, db *sql.DB, cfg *config.ModuleConfig) (any, error) {
			return ent.EntBootstrap(provider, db, func(drv *entsql.Driver) any {
				return generated.NewClient(generated.Driver(drv))
			}, cfg.Database.AutoMigrate)
		},
		OnRegistry: func(app *fiber.App, deps *module.ModuleDeps[config.ModuleConfig]) {
			authSecret := utils.GetAuthSecret()
			db := deps.Ent.(*generated.Client)
			repo := repository.NewTokenRepository(db)
			svc := service.NewTokenService(repo, authSecret)
			h := handler.NewTokenHandler(svc)
			h.RegisterRoutes(app.Group(utils.GetModuleRoute(deps.Config.Server.Route), middleware.RequireInternal()))

			cron.RegisterJob(jobs.NewClearResfreshTokenJob(svc))
		},
	})
}
