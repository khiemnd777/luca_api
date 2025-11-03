package main

import (
	"database/sql"

	entsql "entgo.io/ent/dialect/sql"

	"github.com/gofiber/fiber/v2"
	"github.com/khiemnd777/andy_api/modules/auth/config"
	"github.com/khiemnd777/andy_api/modules/auth/handler"
	"github.com/khiemnd777/andy_api/modules/auth/repository"
	"github.com/khiemnd777/andy_api/modules/auth/service"
	"github.com/khiemnd777/andy_api/shared/db/ent"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/module"
	"github.com/khiemnd777/andy_api/shared/utils"
)

func main() {
	module.StartModule(module.ModuleOptions[config.ModuleConfig]{
		ConfigPath: utils.GetModuleConfigPath("auth"),
		ModuleName: "auth",
		InitEntClient: func(provider string, db *sql.DB, cfg *config.ModuleConfig) (any, error) {
			return ent.EntBootstrap(provider, db, func(drv *entsql.Driver) any {
				return generated.NewClient(generated.Driver(drv))
			}, cfg.Database.AutoMigrate)
		},
		OnRegistry: func(app *fiber.App, deps *module.ModuleDeps[config.ModuleConfig]) {
			authSecret := utils.GetAuthSecret()
			db := deps.Ent.(*generated.Client)
			repo := repository.NewAuthRepository(db)
			svc := service.NewAuthService(repo, authSecret)
			h := handler.NewAuthHandler(svc)
			h.RegisterRoutes(app.Group(utils.GetModuleRoute(deps.Config.Server.Route)))
		},
	})
}
