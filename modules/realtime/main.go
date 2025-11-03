package main

import (
	"database/sql"

	entsql "entgo.io/ent/dialect/sql"
	"github.com/gofiber/fiber/v2"
	"github.com/khiemnd777/andy_api/modules/realtime/config"
	"github.com/khiemnd777/andy_api/shared/db/ent"
	"github.com/khiemnd777/andy_api/shared/middleware"

	"github.com/khiemnd777/andy_api/shared/db/ent/generated"

	"github.com/khiemnd777/andy_api/modules/realtime/handler"
	"github.com/khiemnd777/andy_api/modules/realtime/service"
	"github.com/khiemnd777/andy_api/shared/module"
	"github.com/khiemnd777/andy_api/shared/utils"
)

func main() {
	module.StartModule(module.ModuleOptions[config.ModuleConfig]{
		ConfigPath: utils.GetModuleConfigPath("realtime"),
		ModuleName: "realtime",
		InitEntClient: func(provider string, db *sql.DB, cfg *config.ModuleConfig) (any, error) {

			return ent.EntBootstrap(provider, db, func(drv *entsql.Driver) any {
				return generated.NewClient(generated.Driver(drv))
			}, cfg.Database.AutoMigrate)

		},
		OnRegistry: func(app *fiber.App, deps *module.ModuleDeps[config.ModuleConfig]) {
			hub := service.NewHub()
			hub.InitPubSubEvents()
			go hub.Run()

			jwtSecret := utils.GetAuthSecret()
			h := handler.NewHandler(hub, jwtSecret)
			h.RegisterRoutes(app.Group(utils.GetModuleRoute(deps.Config.Server.Route)))
			h.RegisterInternalRoutes(app.Group(utils.GetModuleRoute(deps.Config.Server.Route), middleware.RequireInternal()))
		},
	})
}
