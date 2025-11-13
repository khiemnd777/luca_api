package main

import (
	"database/sql"

	entsql "entgo.io/ent/dialect/sql"
	"github.com/gofiber/fiber/v2"
	"github.com/khiemnd777/andy_api/modules/search/config"
	"github.com/khiemnd777/andy_api/modules/search/handler"
	"github.com/khiemnd777/andy_api/modules/search/repository"
	"github.com/khiemnd777/andy_api/modules/search/service"
	"github.com/khiemnd777/andy_api/shared/db/ent"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/middleware"
	"github.com/khiemnd777/andy_api/shared/module"
	"github.com/khiemnd777/andy_api/shared/utils"
)

func main() {
	module.StartModule(module.ModuleOptions[config.ModuleConfig]{
		ConfigPath: utils.GetModuleConfigPath("search"),
		ModuleName: "search",
		InitEntClient: func(provider string, db *sql.DB, cfg *config.ModuleConfig) (any, error) {
			return ent.EntBootstrap(provider, db, func(drv *entsql.Driver) any {
				return generated.NewClient(generated.Driver(drv))
			}, cfg.Database.AutoMigrate)
		},
		OnRegistry: func(app *fiber.App, deps *module.ModuleDeps[config.ModuleConfig]) {
			repo := repository.NewSearchRepository(deps)
			svc := service.NewSearchService(repo, deps)
			h := handler.NewSearchHandler(svc, deps)
			h.RegisterRoutes(app.Group(utils.GetModuleRoute(deps.Config.Server.Route), middleware.RequireAuth()))
		},
	})
}
