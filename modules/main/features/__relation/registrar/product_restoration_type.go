package registrar

import (
	policy "github.com/khiemnd777/andy_api/modules/main/features/__relation/policy"
	"github.com/khiemnd777/andy_api/shared/logger"
)

func init() {
	logger.Debug("[RELATION] Register products - restoration types")
	policy.RegisterM2M("products_restoration_types",
		policy.ConfigM2M{
			MainTable:           "products",
			RefTable:            "restoration_types",
			EntityPropMainID:    "ID",
			DTOPropRefIDs:       "RestorationTypeIDs",
			DTOPropDisplayNames: "RestorationTypeNames",

			RefList: &policy.RefListConfig{
				Permissions: []string{"product.view"},
				RefFields:   []string{"id", "category_id", "name"},
				CachePrefix: "restoration_type:list",
			},
		},
	)

	policy.RegisterRefSearch("products_restoration_types", policy.ConfigSearch{
		RefTable:    "restoration_types",
		NormFields:  []string{"name"},
		RefFields:   []string{"id", "category_id", "name"},
		Permissions: []string{"product.search"},
		CachePrefix: "restoration_type:search",
	})
}
