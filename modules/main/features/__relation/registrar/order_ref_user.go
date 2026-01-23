package registrar

import (
	policy "github.com/khiemnd777/andy_api/modules/main/features/__relation/policy"
	"github.com/khiemnd777/andy_api/shared/logger"
	"github.com/khiemnd777/andy_api/shared/utils"
)

func init() {
	logger.Debug("[RELATION] Register order - ref user")
	policy.Register1("orders_ref_users", policy.Config1{
		MainTable:      "orders",
		MainIDProp:     "ID",
		MainRefIDCol:   "ref_user_id",
		MainRefNameCol: utils.Ptr("ref_user_name"),

		RefTable:   "users",
		RefIDCol:   "id",
		RefNameCol: "name",
		RefFields:  []string{"id", "name"},

		UpsertedIDProp:   "RefUserID",
		UpsertedNameProp: utils.Ptr("RefUserName"),

		Permissions: []string{"staff.view"},
		CachePrefix: "staff",
	})
	policy.RegisterRefSearch("orders_ref_users", policy.ConfigSearch{
		RefTable:     "users",
		Alias:        "u",
		NormFields:   []string{"u.name"},
		RefFields:    []string{"id", "name"},
		SelectFields: []string{"u.id", "u.name"},
		Permissions:  []string{"staff.search"},
		CachePrefix:  "staff:search",
	})
}
