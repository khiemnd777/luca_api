package registrar

import (
	policy "github.com/khiemnd777/andy_api/modules/main/features/__relation/policy"
	"github.com/khiemnd777/andy_api/shared/logger"
)

func init() {
	logger.Debug("[RELATION] Register products - brand names")
	policy.RegisterM2M("products_brand_names",
		policy.ConfigM2M{
			MainTable:           "products",
			RefTable:            "brand_names",
			EntityPropMainID:    "ID",
			DTOPropRefIDs:       "BrandNameIDs",
			DTOPropDisplayNames: "BrandNameNames",

			RefList: &policy.RefListConfig{
				Permissions: []string{"product.view"},
				RefFields:   []string{"id", "category_id", "name"},
				CachePrefix: "brand_name:list",
			},
		},
	)

	policy.RegisterRefSearch("products_brand_names", policy.ConfigSearch{
		RefTable:    "brand_names",
		NormFields:  []string{"name"},
		RefFields:   []string{"id", "category_id", "name"},
		Permissions: []string{"product.search"},
		CachePrefix: "brand_name:search",
	})
}
