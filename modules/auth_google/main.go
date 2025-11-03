package main

import (
	"database/sql"

	entsql "entgo.io/ent/dialect/sql"
	"github.com/gofiber/fiber/v2"
	"github.com/khiemnd777/andy_api/shared/db/ent"
	"github.com/khiemnd777/andy_api/shared/utils"

	"github.com/khiemnd777/andy_api/shared/db/ent/generated"

	"github.com/khiemnd777/andy_api/modules/auth_google/config"
	"github.com/khiemnd777/andy_api/modules/auth_google/handler"
	"github.com/khiemnd777/andy_api/modules/auth_google/repository"
	"github.com/khiemnd777/andy_api/modules/auth_google/service"
	"github.com/khiemnd777/andy_api/shared/module"
)

// Manage Google OAuth clients: https://console.cloud.google.com/auth/clients?invt=AbtvKg&project=spry-cat-363606
// keytool -list -v -alias androiddebugkey -keystore "%USERPROFILE%\.android\debug.keystore" -storepass android -keypass android
// ✅ Dành cho Windows
// Nếu Mac/Linux thì đổi %USERPROFILE% → $HOME

func main() {
	module.StartModule(module.ModuleOptions[config.ModuleConfig]{
		ConfigPath: utils.GetModuleConfigPath("auth_google"),
		ModuleName: "auth_google",
		InitEntClient: func(provider string, db *sql.DB, cfg *config.ModuleConfig) (any, error) {
			return ent.EntBootstrap(provider, db, func(drv *entsql.Driver) any {
				return generated.NewClient(generated.Driver(drv))
			}, cfg.Database.AutoMigrate)
		},
		OnRegistry: func(app *fiber.App, deps *module.ModuleDeps[config.ModuleConfig]) {
			authSecret := utils.GetAuthSecret()
			repo := repository.NewAuthGoogleRepository(deps.Ent.(*generated.Client), deps)
			svc := service.NewAuthGoogleService(repo, deps, authSecret)
			h := handler.NewAuthGoogleHandler(svc, deps)
			h.RegisterRoutes(app.Group(utils.GetModuleRoute(deps.Config.Server.Route)))
		},
	})
}
