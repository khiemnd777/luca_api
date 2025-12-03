package registrar

import (
	policy "github.com/khiemnd777/andy_api/modules/main/features/__relation/policy"
	"github.com/khiemnd777/andy_api/shared/logger"
	"github.com/khiemnd777/andy_api/shared/utils"
)

func init() {
	logger.Debug("[RELATION] Register orderitemprocess - assignee")
	policy.Register1("orderitemprocess-assignee", policy.Config1{
		MainTable:      "order_item_processes",
		MainIDProp:     "ID",
		MainRefIDCol:   "assigned_id",
		MainRefNameCol: utils.Ptr("assigned_name"),

		RefTable:   "users",
		RefIDCol:   "id",
		RefNameCol: "name",
		RefFields:  []string{"id", "name"},

		UpsertedIDProp:   "AssignedID",
		UpsertedNameProp: utils.Ptr("AssignedName"),

		Permissions: []string{"staff.view"},
		CachePrefix: "staff",
	})
	policy.RegisterRefSearch("orderitemprocess-assignee", policy.ConfigSearch{
		RefTable:    "users",
		NormFields:  []string{"name"},
		RefFields:   []string{"id", "name"},
		Permissions: []string{"staff.search"},
		CachePrefix: "staff:search",
	})
}
