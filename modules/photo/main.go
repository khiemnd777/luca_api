package main

import (
	"database/sql"

	entsql "entgo.io/ent/dialect/sql"
	"github.com/gofiber/fiber/v2"
	"github.com/khiemnd777/andy_api/modules/photo/config"
	"github.com/khiemnd777/andy_api/modules/photo/jobs"
	"github.com/khiemnd777/andy_api/shared/cron"
	"github.com/khiemnd777/andy_api/shared/db/ent"
	"github.com/khiemnd777/andy_api/shared/middleware"

	"github.com/khiemnd777/andy_api/shared/db/ent/generated"

	"github.com/khiemnd777/andy_api/modules/photo/handler"
	"github.com/khiemnd777/andy_api/modules/photo/repository"
	"github.com/khiemnd777/andy_api/modules/photo/service"
	"github.com/khiemnd777/andy_api/shared/module"
	"github.com/khiemnd777/andy_api/shared/utils"
)

func main() {
	module.StartModule(module.ModuleOptions[config.ModuleConfig]{
		ConfigPath: utils.GetModuleConfigPath("photo"),
		ModuleName: "photo",
		InitEntClient: func(provider string, db *sql.DB, cfg *config.ModuleConfig) (any, error) {
			return ent.EntBootstrap(provider, db, func(drv *entsql.Driver) any {
				return generated.NewClient(generated.Driver(drv))
			}, cfg.Database.AutoMigrate)

		},
		OnRegistry: func(app *fiber.App, deps *module.ModuleDeps[config.ModuleConfig]) {
			repo := repository.NewPhotoRepository(deps.Ent.(*generated.Client), deps)
			svc := service.NewPhotoService(repo, deps)
			h := handler.NewPhotoHandler(svc, deps)
			h.RegisterRoutes(app.Group(utils.GetModuleRoute(deps.Config.Server.Route), middleware.RequireAuth()))

			cron.RegisterJob(jobs.NewClearDeletedPhotoJob(svc))
		},
	})
}
