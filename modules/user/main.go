package main

import (
	"database/sql"

	entsql "entgo.io/ent/dialect/sql"

	"github.com/gofiber/fiber/v2"

	"github.com/khiemnd777/andy_api/shared/db/ent"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/middleware"
	"github.com/khiemnd777/andy_api/shared/utils"

	"github.com/khiemnd777/andy_api/modules/user/config"
	"github.com/khiemnd777/andy_api/modules/user/handler"
	"github.com/khiemnd777/andy_api/modules/user/repository"
	"github.com/khiemnd777/andy_api/modules/user/service"
	"github.com/khiemnd777/andy_api/shared/module"
)

func main() {
	module.StartModule(module.ModuleOptions[config.ModuleConfig]{
		ConfigPath: utils.GetModuleConfigPath("user"),
		ModuleName: "user",
		InitEntClient: func(provider string, db *sql.DB, cfg *config.ModuleConfig) (any, error) {
			return ent.EntBootstrap(provider, db, func(drv *entsql.Driver) any {
				return generated.NewClient(generated.Driver(drv))
			}, cfg.Database.AutoMigrate)
		},
		OnRegistry: func(app *fiber.App, deps *module.ModuleDeps[config.ModuleConfig]) {
			repo := repository.NewUserRepository(deps.Ent.(*generated.Client))
			svc := service.NewUserService(repo, deps)
			h := handler.NewUserHandler(svc, deps)
			h.RegisterRoutes(app.Group(utils.GetModuleRoute(deps.Config.Server.Route), middleware.RequireAuth()))
		},
	})
}
