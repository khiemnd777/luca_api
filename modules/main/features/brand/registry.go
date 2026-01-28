package brand

import (
	"github.com/gofiber/fiber/v2"

	"github.com/khiemnd777/andy_api/modules/main/config"
	"github.com/khiemnd777/andy_api/modules/main/features/brand/handler"
	"github.com/khiemnd777/andy_api/modules/main/features/brand/repository"
	"github.com/khiemnd777/andy_api/modules/main/features/brand/service"
	"github.com/khiemnd777/andy_api/modules/main/registry"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/metadata/customfields"
	"github.com/khiemnd777/andy_api/shared/module"
)

type feature struct{}

func (feature) ID() string    { return "brand" }
func (feature) Priority() int { return 90 }

func (feature) Register(router fiber.Router, deps *module.ModuleDeps[config.ModuleConfig], cfMgr *customfields.Manager) error {
	repo := repository.NewBrandNameRepository(deps.Ent.(*generated.Client), deps)
	svc := service.NewBrandNameService(repo, deps)
	h := handler.NewBrandNameHandler(svc, deps)
	h.RegisterRoutes(router)
	return nil
}

func init() { registry.Register(feature{}) }
