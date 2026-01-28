package registrar

import (
	policy "github.com/khiemnd777/andy_api/modules/main/features/__relation/policy"
	"github.com/khiemnd777/andy_api/shared/logger"
)

func init() {
	logger.Debug("[RELATION] Register products - techniques")
	policy.RegisterM2M("products_techniques",
		policy.ConfigM2M{
			MainTable:           "products",
			RefTable:            "techniques",
			EntityPropMainID:    "ID",
			DTOPropRefIDs:       "TechniqueIDs",
			DTOPropDisplayNames: "TechniqueNames",

			RefList: &policy.RefListConfig{
				Permissions: []string{"product.view"},
				RefFields:   []string{"id", "category_id", "name"},
				CachePrefix: "technique:list",
			},
		},
	)

	policy.RegisterRefSearch("products_techniques", policy.ConfigSearch{
		RefTable:    "techniques",
		NormFields:  []string{"name"},
		RefFields:   []string{"id", "category_id", "name"},
		Permissions: []string{"product.search"},
		CachePrefix: "technique:search",
	})
}
