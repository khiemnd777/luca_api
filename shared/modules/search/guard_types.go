package search

import (
	"github.com/gofiber/fiber/v2"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/modules/search/model"
)

type GuardCtx struct {
	Ctx    *fiber.Ctx
	DB     *generated.Client
	UserID int
	DeptID int
	Perms  map[string]struct{}
}

type Guard func(ctx GuardCtx, rows []model.Row) []model.Row
