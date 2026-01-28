package registrar

import (
	policy "github.com/khiemnd777/andy_api/modules/main/features/__relation/policy"
	"github.com/khiemnd777/andy_api/shared/logger"
)

func init() {
	logger.Debug("[RELATION] Register products - raw materials")
	policy.RegisterM2M("products_raw_materials",
		policy.ConfigM2M{
			MainTable:           "products",
			RefTable:            "raw_materials",
			EntityPropMainID:    "ID",
			DTOPropRefIDs:       "RawMaterialIDs",
			DTOPropDisplayNames: "RawMaterialNames",

			RefList: &policy.RefListConfig{
				Permissions: []string{"product.view"},
				RefFields:   []string{"id", "category_id", "name"},
				CachePrefix: "raw_material:list",
			},
		},
	)

	policy.RegisterRefSearch("products_raw_materials", policy.ConfigSearch{
		RefTable:    "raw_materials",
		NormFields:  []string{"name"},
		RefFields:   []string{"id", "category_id", "name"},
		Permissions: []string{"product.search"},
		CachePrefix: "raw_material:search",
	})
}
