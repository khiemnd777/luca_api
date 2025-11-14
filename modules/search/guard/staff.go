package search

import (
	"fmt"

	"github.com/khiemnd777/andy_api/shared/logger"
	"github.com/khiemnd777/andy_api/shared/middleware/rbac"
	"github.com/khiemnd777/andy_api/shared/modules/search"
	"github.com/khiemnd777/andy_api/shared/modules/search/model"
)

func init() {
	logger.Debug("[GuardSearch] Register Staff")
	search.RegisterGuard("staff", func(ctx search.GuardCtx, rows []model.Row) []model.Row {
		perms := ctx.Perms

		logger.Debug(fmt.Sprintf("[GuardSearch] perms: %v", perms))

		if !rbac.HasAnyPerm(perms, "staff.search") {
			return []model.Row{}
		}

		out := make([]model.Row, 0, len(rows))
		out = append(out, rows...)

		return out
	})
}
