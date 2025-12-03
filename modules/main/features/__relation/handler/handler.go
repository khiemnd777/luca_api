package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/khiemnd777/andy_api/modules/main/config"
	relation "github.com/khiemnd777/andy_api/modules/main/features/__relation/policy"
	"github.com/khiemnd777/andy_api/modules/main/features/__relation/service"
	"github.com/khiemnd777/andy_api/shared/app"
	"github.com/khiemnd777/andy_api/shared/app/client_error"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	dbutils "github.com/khiemnd777/andy_api/shared/db/utils"
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
	app.RouterGet(router, "/:dept_id<int>/relation/:key/:main_id<int>/one", h.Get1)
	app.RouterGet(router, "/:dept_id<int>/relation/:key/:main_id<int>/1n/list", h.List1N)
	app.RouterGet(router, "/:dept_id<int>/relation/:key/:main_id<int>/m2m/list", h.ListM2M)
	app.RouterGet(router, "/:dept_id<int>/relation/:key/search", h.Search)
}

func (h *RelationHandler) Get1(c *fiber.Ctx) error {
	key := c.Params("key")
	if key == "" {
		return client_error.ResponseError(c, fiber.StatusBadRequest, nil, "missing key or ref")
	}

	mainID, err := utils.GetParamAsInt(c, "main_id")
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, err.Error())
	}

	cfg, err := relation.GetConfig1(key)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusNotFound, err, err.Error())
	}

	if len(cfg.Permissions) > 0 {
		if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), cfg.Permissions...); err != nil {
			return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
		}
	}

	res, err := h.svc.Get1(c.UserContext(), key, mainID)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}

	return c.JSON(res)
}

func (h *RelationHandler) List1N(c *fiber.Ctx) error {
	key := c.Params("key")
	if key == "" {
		return client_error.ResponseError(c, fiber.StatusBadRequest, nil, "missing key or ref")
	}

	mainID, err := utils.GetParamAsInt(c, "main_id")
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, err.Error())
	}

	cfg, err := relation.GetConfig1N(key)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusNotFound, err, err.Error())
	}

	if len(cfg.Permissions) > 0 {
		if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), cfg.Permissions...); err != nil {
			return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
		}
	}

	q := table.ParseTableQuery(c, 10)

	res, err := h.svc.List1N(c.UserContext(), key, mainID, q)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}

	return c.JSON(res)
}

func (h *RelationHandler) ListM2M(c *fiber.Ctx) error {
	key := c.Params("key")
	if key == "" {
		return client_error.ResponseError(c, fiber.StatusBadRequest, nil, "missing key or ref")
	}

	mainID, err := utils.GetParamAsInt(c, "main_id")
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, err.Error())
	}

	cfg, err := relation.GetConfigM2M(key)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusNotFound, err, err.Error())
	}

	if cfg.GetRefList != nil && len(cfg.GetRefList.Permissions) > 0 {
		if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), cfg.GetRefList.Permissions...); err != nil {
			return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
		}
	}

	q := table.ParseTableQuery(c, 10)

	res, err := h.svc.ListM2M(c.UserContext(), key, mainID, q)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}

	return c.JSON(res)
}

func (h *RelationHandler) Search(c *fiber.Ctx) error {
	key := c.Params("key")
	if key == "" {
		return client_error.ResponseError(c, fiber.StatusBadRequest, nil, "missing key or ref")
	}

	cfg, err := relation.GetConfigRefSearch(key)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusNotFound, err, err.Error())
	}

	if len(cfg.Permissions) > 0 {
		if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), cfg.Permissions...); err != nil {
			return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
		}
	}

	q := dbutils.ParseSearchQuery(c, 10)

	res, err := h.svc.Search(c.UserContext(), key, q)

	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}

	return c.JSON(res)
}
