package main

import (
	"database/sql"

	entsql "entgo.io/ent/dialect/sql"
	"github.com/gofiber/fiber/v2"
	"github.com/khiemnd777/andy_api/modules/main/config"
	"github.com/khiemnd777/andy_api/modules/main/department/handler"
	"github.com/khiemnd777/andy_api/modules/main/department/repository"
	"github.com/khiemnd777/andy_api/modules/main/department/service"
	"github.com/khiemnd777/andy_api/shared/db/ent"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/middleware"
	"github.com/khiemnd777/andy_api/shared/module"
	"github.com/khiemnd777/andy_api/shared/utils"
)

func main() {
	module.StartModule(module.ModuleOptions[config.ModuleConfig]{
		ConfigPath: utils.GetModuleConfigPath("main"),
		ModuleName: "main",
		InitEntClient: func(provider string, db *sql.DB, cfg *config.ModuleConfig) (any, error) {
			return ent.EntBootstrap(provider, db, func(drv *entsql.Driver) any {
				return generated.NewClient(generated.Driver(drv))
			}, cfg.Database.AutoMigrate)
		},
		OnRegistry: func(app *fiber.App, deps *module.ModuleDeps[config.ModuleConfig]) {
			repo := repository.NewDepartmentRepository(deps.Ent.(*generated.Client), deps)
			svc := service.NewDepartmentService(repo, deps)
			svcDeptGuard := service.NewGuardService(repo)

			groupRouter := app.Group(utils.GetModuleRoute(deps.Config.Server.Route), middleware.RequireAuth())

			groupRouter.Use("/:dept_id<int>/*",
				middleware.RequireDepartmentMember(svcDeptGuard, "dept_id"),
			)

			h := handler.NewDepartmentHandler(svc, deps)
			h.RegisterRoutes(groupRouter)
		},
	})
}
