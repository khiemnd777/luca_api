package material

import (
	"github.com/gofiber/fiber/v2"

	"github.com/khiemnd777/andy_api/modules/main/config"
	"github.com/khiemnd777/andy_api/modules/main/features/material/handler"
	"github.com/khiemnd777/andy_api/modules/main/features/material/repository"
	"github.com/khiemnd777/andy_api/modules/main/features/material/service"
	"github.com/khiemnd777/andy_api/modules/main/registry"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/metadata/customfields"
	"github.com/khiemnd777/andy_api/shared/module"
)

type feature struct{}

func (feature) ID() string    { return "material" }
func (feature) Priority() int { return 60 }

func (feature) Register(router fiber.Router, deps *module.ModuleDeps[config.ModuleConfig], cfMgr *customfields.Manager) error {
	repo := repository.NewMaterialRepository(deps.Ent.(*generated.Client), deps, cfMgr)
	svc := service.NewMaterialService(repo, deps, cfMgr)
	h := handler.NewMaterialHandler(svc, deps)
	h.RegisterRoutes(router)
	return nil
}

func init() { registry.Register(feature{}) }
