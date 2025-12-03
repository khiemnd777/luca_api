package registrar

import (
	policy "github.com/khiemnd777/andy_api/modules/main/features/__relation/policy"
	"github.com/khiemnd777/andy_api/shared/logger"
)

func init() {
	logger.Debug("[RELATION] Register materials - suppliers")
	policy.RegisterM2M("material",
		policy.ConfigM2M{
			MainTable:   "materials",
			RefTable:    "suppliers",
			MainIDProp:  "ID",
			RefIDsProp:  "SupplierIDs",
			DisplayProp: "SupplierNames",

			RefList: &policy.RefListConfig{
				Permissions: []string{"supplier.view"},
				RefFields:   []string{"id", "code", "name"},
				CachePrefix: "supplier:list",
			},
		},
	)
	policy.RegisterRefSearch("material", policy.ConfigSearch{
		RefTable:    "suppliers",
		NormFields:  []string{"code", "name"},
		RefFields:   []string{"id", "code", "name"},
		Permissions: []string{"supplier.search"},
		CachePrefix: "supplier:list",
	})
}
