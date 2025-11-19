package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/khiemnd777/andy_api/modules/main/config"
	relation "github.com/khiemnd777/andy_api/modules/main/features/__relation/policy"
	"github.com/khiemnd777/andy_api/modules/main/features/__relation/service"
	"github.com/khiemnd777/andy_api/shared/app"
	"github.com/khiemnd777/andy_api/shared/app/client_error"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/middleware/rbac"
	"github.com/khiemnd777/andy_api/shared/module"
	"github.com/khiemnd777/andy_api/shared/utils"
	"github.com/khiemnd777/andy_api/shared/utils/table"
)

type RelationHandler struct {
	svc  *service.RelationService
	deps *module.ModuleDeps[config.ModuleConfig]
}

func NewRelationHandler(svc *service.RelationService, deps *module.ModuleDeps[config.ModuleConfig]) *RelationHandler {
	return &RelationHandler{svc: svc, deps: deps}
}

func (h *RelationHandler) RegisterRoutes(router fiber.Router) {
	app.RouterGet(router, "/:dept_id<int>/relation/:key/:main_id<int>/list", h.List)
}

// GET /relation/:main/:main_id/:ref/list
func (h *RelationHandler) List(c *fiber.Ctx) error {

	key := c.Params("key")

	if key == "" {
		return client_error.ResponseError(c, fiber.StatusBadRequest, nil, "missing key or ref")
	}

	mainID, err := utils.GetParamAsInt(c, "main_id")
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, err.Error())
	}

	cfg, err := relation.GetConfig(key)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusNotFound, err, err.Error())
	}

	if cfg.GetRefList != nil && len(cfg.GetRefList.Permissions) > 0 {
		if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), cfg.GetRefList.Permissions...); err != nil {
			return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
		}
	}

	q := table.ParseTableQuery(c, 10)

	res, err := h.svc.List(c.UserContext(), key, mainID, q)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}

	return c.JSON(res)
}
