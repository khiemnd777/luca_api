package order

import (
	"github.com/gofiber/fiber/v2"

	"github.com/khiemnd777/andy_api/modules/main/config"
	"github.com/khiemnd777/andy_api/modules/main/features/order/handler"
	"github.com/khiemnd777/andy_api/modules/main/features/order/jobs"
	"github.com/khiemnd777/andy_api/modules/main/features/order/repository"
	"github.com/khiemnd777/andy_api/modules/main/features/order/service"
	"github.com/khiemnd777/andy_api/modules/main/registry"
	"github.com/khiemnd777/andy_api/shared/cron"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/metadata/customfields"
	"github.com/khiemnd777/andy_api/shared/module"
)

type feature struct{}

func (feature) ID() string    { return "order" }
func (feature) Priority() int { return 79 }

func (feature) Register(router fiber.Router, deps *module.ModuleDeps[config.ModuleConfig], cfMgr *customfields.Manager) error {
	ordRepo := repository.NewOrderRepository(deps.Ent.(*generated.Client), deps, cfMgr)
	ordSvc := service.NewOrderService(ordRepo, deps, cfMgr)
	ordHandler := handler.NewOrderHandler(ordSvc, deps)
	ordHandler.RegisterRoutes(router)

	orderCodeSvc := service.NewOrderCodeService(deps.Ent.(*generated.Client))
	cron.RegisterJob(jobs.NewClearExpiredOrderCodeJob(orderCodeSvc))
	cron.RegisterJob(jobs.NewExpireOrderCodeJob(orderCodeSvc))
	orderCodeHandler := handler.NewOrderCodeHandler(orderCodeSvc, deps)
	orderCodeHandler.RegisterRoutes(router)

	ordItemRepo := repository.NewOrderItemRepository(deps.Ent.(*generated.Client), deps, cfMgr)
	ordItemSvc := service.NewOrderItemService(ordItemRepo, deps, cfMgr)
	ordItemHandler := handler.NewOrderItemHandler(ordItemSvc, deps)
	ordItemHandler.RegisterRoutes(router)

	ordItemMaterialRepo := repository.NewOrderItemMaterialRepository(deps.Ent.(*generated.Client))
	ordItemMaterialSvc := service.NewOrderItemMaterialService(ordItemMaterialRepo, deps)
	ordItemMaterialHandler := handler.NewOrderItemMaterialHandler(ordItemMaterialSvc, deps)
	ordItemMaterialHandler.RegisterRoutes(router)

	ordItemProcessRepo := repository.NewOrderItemProcessRepository(deps.Ent.(*generated.Client), deps, cfMgr)
	ordItemProcessInProgressRepo := repository.NewOrderItemProcessInProgressRepository(deps.Ent.(*generated.Client), ordItemProcessRepo)
	ordItemProcessSvc := service.NewOrderItemProcessService(ordItemProcessRepo, ordItemProcessInProgressRepo, deps, cfMgr)
	ordItemProcessHandler := handler.NewOrderItemProcessHandler(ordItemProcessSvc, deps)
	ordItemProcessHandler.RegisterRoutes(router)

	return nil
}

func init() { registry.Register(feature{}) }
