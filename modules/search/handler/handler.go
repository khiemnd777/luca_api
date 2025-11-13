package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/khiemnd777/andy_api/modules/search/config"
	"github.com/khiemnd777/andy_api/modules/search/model"
	"github.com/khiemnd777/andy_api/modules/search/service"
	"github.com/khiemnd777/andy_api/shared/app"
	"github.com/khiemnd777/andy_api/shared/app/client_error"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/module"
	"github.com/khiemnd777/andy_api/shared/modules/search"
	"github.com/khiemnd777/andy_api/shared/utils"
)

type SearchHandler struct {
	svc  service.SearchService
	deps *module.ModuleDeps[config.ModuleConfig]
}

func NewSearchHandler(svc service.SearchService, deps *module.ModuleDeps[config.ModuleConfig]) *SearchHandler {
	return &SearchHandler{svc: svc, deps: deps}
}

func (h *SearchHandler) RegisterRoutes(router fiber.Router) {
	app.RouterGet(router, "/", h.Search)
}

func (h *SearchHandler) Search(c *fiber.Ctx) error {
	q := utils.GetQueryAsString(c, "q")
	deptID, _ := utils.GetDeptIDInt(c)

	rows, err := h.svc.Search(c.UserContext(), model.Options{
		Query:           q,
		OrgID:           utils.Ptr(int64(deptID)),
		UseTrgmFallback: true,
	})

	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}

	filtered := search.GuardSearch(c, h.deps.Ent.(*generated.Client), rows)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"items": filtered,
		"total": len(filtered),
	})
}
