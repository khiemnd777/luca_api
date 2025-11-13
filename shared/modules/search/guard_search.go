package search

import (
	"github.com/gofiber/fiber/v2"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/modules/search/model"
	"github.com/khiemnd777/andy_api/shared/utils"
)

func GuardSearch(c *fiber.Ctx, dbEnt *generated.Client, in []model.Row) []model.Row {
	if len(in) == 0 {
		return in
	}

	userID, _ := utils.GetUserIDInt(c)
	deptID, _ := utils.GetDeptIDInt(c)
	perms, _ := utils.GetPermSetFromClaims(c)

	ctx := GuardCtx{
		Ctx:    c,
		DB:     dbEnt,
		UserID: userID,
		DeptID: deptID,
		Perms:  perms,
	}

	buckets := map[string][]model.Row{}
	order := make([]string, 0, 8)
	for _, r := range in {
		if _, ok := buckets[r.EntityType]; !ok {
			order = append(order, r.EntityType)
		}
		buckets[r.EntityType] = append(buckets[r.EntityType], r)
	}

	out := make([]model.Row, 0, len(in))
	for _, t := range order {
		if g, ok := guardRegistry[t]; ok {
			out = append(out, g(ctx, buckets[t])...)
		} else {
			out = append(out, buckets[t]...)
		}
	}
	return out
}
