package promotion

import (
	"github.com/gofiber/fiber/v2"

	"github.com/khiemnd777/andy_api/modules/main/config"
	"github.com/khiemnd777/andy_api/modules/main/features/order/repository"
	orderservice "github.com/khiemnd777/andy_api/modules/main/features/order/service"
	"github.com/khiemnd777/andy_api/modules/main/features/promotion/handler"
	promotionrepo "github.com/khiemnd777/andy_api/modules/main/features/promotion/repository"
	promotionservice "github.com/khiemnd777/andy_api/modules/main/features/promotion/service"
	"github.com/khiemnd777/andy_api/modules/main/registry"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/metadata/customfields"
	"github.com/khiemnd777/andy_api/shared/module"
)

type feature struct{}

func (feature) ID() string    { return "promotion" }
func (feature) Priority() int { return 70 }

func (feature) Register(router fiber.Router, deps *module.ModuleDeps[config.ModuleConfig], cfMgr *customfields.Manager) error {
	entClient := deps.Ent.(*generated.Client)
	repo := promotionrepo.NewPromotionRepository(entClient, deps.DB)
	svc := promotionservice.NewPromotionService(repo, deps)

	orderRepo := repository.NewOrderRepository(entClient, deps, cfMgr)
	orderSvc := orderservice.NewOrderService(orderRepo, deps, cfMgr)

	h := handler.NewPromotionHandler(svc, orderSvc, deps)
	h.RegisterRoutes(router)

	adminHandler := handler.NewPromotionAdminHandler(svc, deps)
	adminHandler.RegisterRoutes(router)
	return nil
}

func init() { registry.Register(feature{}) }
