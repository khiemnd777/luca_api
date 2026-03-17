package handler

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/khiemnd777/andy_api/modules/observability/config"
	"github.com/khiemnd777/andy_api/modules/observability/model"
	"github.com/khiemnd777/andy_api/modules/observability/service"
	"github.com/khiemnd777/andy_api/shared/app"
	"github.com/khiemnd777/andy_api/shared/app/client_error"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/middleware/rbac"
	"github.com/khiemnd777/andy_api/shared/module"
)

type ObservabilityHandler struct {
	svc  service.ObservabilityService
	deps *module.ModuleDeps[config.ModuleConfig]
}

func NewObservabilityHandler(svc service.ObservabilityService, deps *module.ModuleDeps[config.ModuleConfig]) *ObservabilityHandler {
	return &ObservabilityHandler{svc: svc, deps: deps}
}

func (h *ObservabilityHandler) RegisterRoutes(router fiber.Router) {
	app.RouterGet(router, "/logs", h.ListLogs)
	app.RouterGet(router, "/logs/summary", h.GetSummary)
}

func (h *ObservabilityHandler) ListLogs(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), "system_log.read"); err != nil {
		return err
	}

	query, err := parseListLogsQuery(c)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, err.Error())
	}

	items, err := h.svc.ListLogs(c.UserContext(), query)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, "Failed to fetch system logs")
	}

	return c.JSON(fiber.Map{
		"items": items,
		"total": len(items),
	})
}

func (h *ObservabilityHandler) GetSummary(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), "system_log.read"); err != nil {
		return err
	}

	query, err := parseListLogsQuery(c)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, err.Error())
	}

	summary, err := h.svc.GetSummary(c.UserContext(), query)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, "Failed to fetch system log summary")
	}

	return c.JSON(summary)
}

func parseListLogsQuery(c *fiber.Ctx) (model.ListLogsQuery, error) {
	from, err := parseTimeQuery(c.Query("from"))
	if err != nil {
		return model.ListLogsQuery{}, err
	}
	to, err := parseTimeQuery(c.Query("to"))
	if err != nil {
		return model.ListLogsQuery{}, err
	}

	levels := splitCSV(c.Query("level"))
	if len(levels) == 0 {
		levels = splitCSV(c.Query("levels"))
	}

	return model.ListLogsQuery{
		Levels:    levels,
		Module:    c.Query("module"),
		Service:   c.Query("service"),
		RequestID: c.Query("request_id"),
		Keyword:   c.Query("keyword"),
		Direction: c.Query("direction"),
		From:      from,
		To:        to,
		Limit:     c.QueryInt("limit", 50),
	}, nil
}

func parseTimeQuery(raw string) (time.Time, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}, nil
	}

	if unixNanos, err := time.Parse(time.RFC3339Nano, raw); err == nil {
		return unixNanos, nil
	}

	if unixSeconds, err := time.Parse(time.RFC3339, raw); err == nil {
		return unixSeconds, nil
	}

	return time.Time{}, fiber.NewError(fiber.StatusBadRequest, "invalid time format, expected RFC3339")
}

func splitCSV(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}

	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}
