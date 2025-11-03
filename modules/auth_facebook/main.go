package main

import (
	"database/sql"

	entsql "entgo.io/ent/dialect/sql"
	"github.com/gofiber/fiber/v2"
	"github.com/khiemnd777/andy_api/shared/db/ent"
	"github.com/khiemnd777/andy_api/shared/utils"

	"github.com/khiemnd777/andy_api/shared/db/ent/generated"

	"github.com/khiemnd777/andy_api/modules/auth_facebook/config"
	"github.com/khiemnd777/andy_api/modules/auth_facebook/handler"
	"github.com/khiemnd777/andy_api/modules/auth_facebook/repository"
	"github.com/khiemnd777/andy_api/modules/auth_facebook/service"
	"github.com/khiemnd777/andy_api/shared/module"
)

// Config Facebook OAuth: https://developers.facebook.com/apps/1209389080651251/settings/basic/

func main() {
	module.StartModule(module.ModuleOptions[config.ModuleConfig]{
		ConfigPath: utils.GetModuleConfigPath("auth_facebook"),
		ModuleName: "auth_facebook",
		InitEntClient: func(provider string, db *sql.DB, cfg *config.ModuleConfig) (any, error) {
			return ent.EntBootstrap(provider, db, func(drv *entsql.Driver) any {
				return generated.NewClient(generated.Driver(drv))
			}, cfg.Database.AutoMigrate)
		},
		OnRegistry: func(app *fiber.App, deps *module.ModuleDeps[config.ModuleConfig]) {
			authSecret := utils.GetAuthSecret()
			repo := repository.NewAuthFacebookRepository(deps.Ent.(*generated.Client), deps)
			svc := service.NewAuthFacebookService(repo, deps, authSecret)
			h := handler.NewAuthFacebookHandler(svc, deps)
			h.RegisterRoutes(app.Group(utils.GetModuleRoute(deps.Config.Server.Route)))
		},
	})
}
