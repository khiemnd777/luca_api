package search

import (
	"github.com/khiemnd777/andy_api/shared/logger"
	"github.com/khiemnd777/andy_api/shared/middleware/rbac"
	"github.com/khiemnd777/andy_api/shared/modules/search"
	"github.com/khiemnd777/andy_api/shared/modules/search/model"
)

func init() {
	logger.Debug("[GuardSearch] Register Customer")
	search.RegisterGuard("customer", func(ctx search.GuardCtx, rows []model.Row) []model.Row {
		perms := ctx.Perms

		if !rbac.HasAnyPerm(perms, "customer.search") {
			return []model.Row{}
		}

		out := make([]model.Row, 0, len(rows))
		out = append(out, rows...)

		return out
	})
}
