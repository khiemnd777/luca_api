package registrar

import (
	model "github.com/khiemnd777/andy_api/modules/main/features/__model"
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

		RefTable:   "staffs",
		RefIDCol:   "id",
		RefNameCol: "name",
		RefDTO:     model.StaffShortDTO{},

		UpsertedIDProp:   "AssignedID",
		UpsertedNameProp: utils.Ptr("AssignedName"),

		Permissions: []string{"staff.view"},
		CachePrefix: "staff",
	})
	policy.RegisterRefSearch("orderitemprocess-assignee", policy.ConfigSearch{
		RefTable:    "staffs",
		NormFields:  []string{"name"},
		Permissions: []string{"staff.search"},
		RefDTO:      model.StaffShortDTO{},
		CachePrefix: "staff:search",
	})
}
