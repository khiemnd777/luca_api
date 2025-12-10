package category

import (
	"github.com/gofiber/fiber/v2"

	"github.com/khiemnd777/andy_api/modules/main/config"
	"github.com/khiemnd777/andy_api/modules/main/features/category/handler"
	"github.com/khiemnd777/andy_api/modules/main/features/category/repository"
	"github.com/khiemnd777/andy_api/modules/main/features/category/service"
	"github.com/khiemnd777/andy_api/modules/main/registry"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/metadata/customfields"
	"github.com/khiemnd777/andy_api/shared/module"
)

type feature struct{}

func (feature) ID() string    { return "category" }
func (feature) Priority() int { return 90 }

func (feature) Register(router fiber.Router, deps *module.ModuleDeps[config.ModuleConfig], cfMgr *customfields.Manager) error {
	repo := repository.NewCategoryRepository(deps.Ent.(*generated.Client), deps, cfMgr)
	svc := service.NewCategoryService(repo, deps, cfMgr)
	h := handler.NewCategoryHandler(svc, deps)
	h.RegisterRoutes(router)
	return nil
}

func init() { registry.Register(feature{}) }
