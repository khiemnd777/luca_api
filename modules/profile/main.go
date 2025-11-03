// modules/profile/main.go
package main

import (
	"database/sql"

	entsql "entgo.io/ent/dialect/sql"

	"github.com/gofiber/fiber/v2"

	"github.com/khiemnd777/andy_api/modules/profile/config"
	"github.com/khiemnd777/andy_api/modules/profile/handler"
	"github.com/khiemnd777/andy_api/modules/profile/repository"
	"github.com/khiemnd777/andy_api/modules/profile/service"
	"github.com/khiemnd777/andy_api/shared/db/ent"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/middleware"
	"github.com/khiemnd777/andy_api/shared/module"
	"github.com/khiemnd777/andy_api/shared/utils"
)

func main() {
	module.StartModule(module.ModuleOptions[config.ModuleConfig]{
		ConfigPath: utils.GetModuleConfigPath("profile"),
		ModuleName: "profile",
		InitEntClient: func(provider string, db *sql.DB, cfg *config.ModuleConfig) (any, error) {
			return ent.InitEntClient(provider, db, func(drv *entsql.Driver) any {
				return generated.NewClient(generated.Driver(drv))
			})
		},
		OnRegistry: func(app *fiber.App, deps *module.ModuleDeps[config.ModuleConfig]) {
			db := deps.Ent.(*generated.Client)
			repo := repository.NewProfileRepository(db, deps)
			svc := service.NewProfileService(repo, deps)
			h := handler.NewProfileHandler(svc, deps)
			h.RegisterRoutes(app.Group(utils.GetModuleRoute(deps.Config.Server.Route), middleware.RequireAuth()))
		},
	})
}
