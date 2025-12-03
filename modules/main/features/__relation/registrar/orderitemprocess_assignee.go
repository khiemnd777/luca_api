package registrar

import (
	"fmt"

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
		RefTable:     "staffs",
		Alias:        "s",
		NormFields:   []string{"u.name"},
		RefFields:    []string{"id", "name"},
		SelectFields: []string{"u.id", "u.name"},
		Permissions:  []string{"staff.search"},
		CachePrefix:  "staff:search",
		ExtraJoins: func() string {
			return `
				JOIN users u ON u.id = s.user_staff
				JOIN user_roles ur ON ur.user_id = u.id
				JOIN roles r ON r.id = ur.role_id
			`
		},
		ExtraWhere: func(args *[]any) string {
			*args = append(*args, "technician")
			return fmt.Sprintf("r.role_name = $%d", len(*args))
		},
	})
}
