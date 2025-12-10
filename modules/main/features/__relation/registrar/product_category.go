package registrar

import (
	policy "github.com/khiemnd777/andy_api/modules/main/features/__relation/policy"
	"github.com/khiemnd777/andy_api/shared/logger"
	"github.com/khiemnd777/andy_api/shared/utils"
)

func init() {
	logger.Debug("[RELATION] Register product - category")
	policy.Register1("product_category", policy.Config1{
		MainTable:      "product",
		MainIDProp:     "ID",
		MainRefIDCol:   "category_id",
		MainRefNameCol: utils.Ptr("category_name"),

		RefNameCol:       "name",
		UpsertedIDProp:   "CategoryID",
		UpsertedNameProp: utils.Ptr("CategoryName"),

		// Get1
		RefTable:    "categories",
		RefIDCol:    "id",
		RefFields:   []string{"id", "name"},
		Permissions: []string{"product.view"},
		CachePrefix: "category",
	})

	policy.RegisterRefSearch("product_category", policy.ConfigSearch{
		RefTable:     "categories",
		NormFields:   []string{"r.name"},
		RefFields:    []string{"id", "name", "parent_id"},
		SelectFields: []string{"r.id", "r.name", "r.parent_id"},
		Permissions:  []string{"product.search"},
		CachePrefix:  "category:search",
	})
}
