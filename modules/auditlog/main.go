package main

import (
	"database/sql"

	entsql "entgo.io/ent/dialect/sql"

	"github.com/gofiber/fiber/v2"

	"github.com/khiemnd777/andy_api/shared/db/ent"
	"github.com/khiemnd777/andy_api/shared/middleware"
	"github.com/khiemnd777/andy_api/shared/module"
	"github.com/khiemnd777/andy_api/shared/utils"

	"github.com/khiemnd777/andy_api/modules/auditlog/config"
	"github.com/khiemnd777/andy_api/modules/auditlog/ent/bootstrap"
	"github.com/khiemnd777/andy_api/modules/auditlog/ent/generated"
	"github.com/khiemnd777/andy_api/modules/auditlog/handler"
	"github.com/khiemnd777/andy_api/modules/auditlog/repository"
	"github.com/khiemnd777/andy_api/modules/auditlog/service"
	sharedGenerated "github.com/khiemnd777/andy_api/shared/db/ent/generated"
)

func main() {
	module.StartModule(module.ModuleOptions[config.ModuleConfig]{
		ConfigPath: utils.GetModuleConfigPath("auditlog"),
		ModuleName: "auditlog",
		InitEntClient: func(provider string, db *sql.DB, cfg *config.ModuleConfig) (any, error) {
			return bootstrap.EntBootstrap(provider, db, func(drv *entsql.Driver) any {
				return generated.NewClient(generated.Driver(drv))
			}, cfg.Database.AutoMigrate)
		},
		InitSharedEntClient: func(provider string, db *sql.DB, cfg *config.ModuleConfig) (any, error) {
			return ent.EntBootstrap(provider, db, func(drv *entsql.Driver) any {
				return sharedGenerated.NewClient(sharedGenerated.Driver(drv))
			}, cfg.SharedDatabase.AutoMigrate)
		},
		OnRegistry: func(app *fiber.App, deps *module.ModuleDeps[config.ModuleConfig]) {
			repo := repository.NewAuditLogRepository(deps.Ent.(*generated.Client), deps.SharedEnt.(*sharedGenerated.Client), deps)
			svc := service.NewAuditLogService(repo, deps)
			svc.InitPubSubEvents()
			h := handler.NewAuditLogHandler(svc)
			h.RegisterRoutes(app.Group(utils.GetModuleRoute(deps.Config.Server.Route), middleware.RequireAuth()))
		},
	})
}
