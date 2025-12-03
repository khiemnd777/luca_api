package registrar

import (
	policy "github.com/khiemnd777/andy_api/modules/main/features/__relation/policy"
	"github.com/khiemnd777/andy_api/shared/logger"
	"github.com/khiemnd777/andy_api/shared/utils"
)

func init() {
	logger.Debug("[RELATION] Register order - product")
	policy.Register1("orders_products", policy.Config1{
		MainTable:      "order_items",
		MainIDProp:     "ID",
		MainRefIDCol:   "product_id",
		MainRefNameCol: utils.Ptr("product_name"),

		RefTable:   "products",
		RefIDCol:   "id",
		RefNameCol: "name",
		RefFields:  []string{"id", "name", "code", "custom_fields"},

		UpsertedIDProp:   "ProductID",
		UpsertedNameProp: utils.Ptr("ProductName"),

		Permissions: []string{"product.view"},
		CachePrefix: "product",
	})
	policy.RegisterRefSearch("orders_products", policy.ConfigSearch{
		RefTable:    "products",
		NormFields:  []string{"code", "name"},
		RefFields:   []string{"id", "name", "code"},
		Permissions: []string{"product.search"},
		CachePrefix: "product:search",
	})
}
