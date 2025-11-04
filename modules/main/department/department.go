package department

import (
	"github.com/gofiber/fiber/v2"
	"github.com/khiemnd777/andy_api/modules/main/config"
	"github.com/khiemnd777/andy_api/modules/main/department/handler"
	"github.com/khiemnd777/andy_api/modules/main/department/repository"
	"github.com/khiemnd777/andy_api/modules/main/department/service"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/module"
)

func NewDepartmentModule(app *fiber.App, router fiber.Router, deps *module.ModuleDeps[config.ModuleConfig]) {
	repo := repository.NewDepartmentRepository(deps.Ent.(*generated.Client), deps)
	svc := service.NewDepartmentService(repo, deps)
	h := handler.NewDepartmentHandler(svc, deps)
	h.RegisterRoutes(router)
}
