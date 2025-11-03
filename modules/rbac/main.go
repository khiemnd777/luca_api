package main

import (
	"database/sql"

	entsql "entgo.io/ent/dialect/sql"

	"github.com/gofiber/fiber/v2"

	"github.com/khiemnd777/andy_api/modules/rbac/config"
	"github.com/khiemnd777/andy_api/modules/rbac/handler"
	"github.com/khiemnd777/andy_api/modules/rbac/repository"
	"github.com/khiemnd777/andy_api/modules/rbac/service"
	"github.com/khiemnd777/andy_api/shared/db/ent"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/middleware"
	"github.com/khiemnd777/andy_api/shared/module"
	"github.com/khiemnd777/andy_api/shared/utils"
)

func main() {
	module.StartModule(module.ModuleOptions[config.ModuleConfig]{
		ConfigPath: utils.GetModuleConfigPath("rbac"),
		ModuleName: "rbac",
		InitEntClient: func(provider string, db *sql.DB, cfg *config.ModuleConfig) (any, error) {
			return ent.EntBootstrap(provider, db, func(drv *entsql.Driver) any {
				return generated.NewClient(generated.Driver(drv))
			}, cfg.Database.AutoMigrate)
		},
		OnRegistry: func(app *fiber.App, deps *module.ModuleDeps[config.ModuleConfig]) {
			db := deps.Ent.(*generated.Client)

			rolesRepo := repository.NewRoleRepository(db)
			permsRepo := repository.NewPermissionRepository(db)
			svc := service.NewRBACService(rolesRepo, permsRepo)
			h := handler.NewRBACHandler(db, svc)
			h.RegisterRoutes(app.Group(utils.GetModuleRoute(deps.Config.Server.Route), middleware.RequireAuth()))
		},
	})
}
