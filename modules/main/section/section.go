package section

import (
	"github.com/gofiber/fiber/v2"
	"github.com/khiemnd777/andy_api/modules/main/config"
	"github.com/khiemnd777/andy_api/modules/main/section/handler"
	"github.com/khiemnd777/andy_api/modules/main/section/repository"
	"github.com/khiemnd777/andy_api/modules/main/section/service"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/module"
)

func NewSection(deps *module.ModuleDeps[config.ModuleConfig], router fiber.Router) {
	repo := repository.NewSectionRepository(deps.Ent.(*generated.Client), deps)
	svc := service.NewSectionService(repo, deps)
	h := handler.NewSectionHandler(svc, deps)
	h.RegisterRoutes(router)
}
