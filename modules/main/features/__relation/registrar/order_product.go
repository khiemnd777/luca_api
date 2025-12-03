package registrar

import (
	model "github.com/khiemnd777/andy_api/modules/main/features/__model"
	policy "github.com/khiemnd777/andy_api/modules/main/features/__relation/policy"
	"github.com/khiemnd777/andy_api/shared/logger"
	"github.com/khiemnd777/andy_api/shared/utils"
)

func init() {
	logger.Debug("[RELATION] Register order - product")
	policy.Register1("order-product", policy.Config1{
		MainTable:      "order_items",
		MainIDProp:     "ID",
		MainRefIDCol:   "product_id",
		MainRefNameCol: utils.Ptr("product_name"),

		RefTable:   "products",
		RefIDCol:   "id",
		RefNameCol: "name",
		RefDTO:     model.ProductShortDTO{},

		UpsertedIDProp:   "ProductID",
		UpsertedNameProp: utils.Ptr("ProductName"),

		Permissions: []string{"product.view"},
		CachePrefix: "product",
	})
	policy.RegisterRefSearch("order-product", policy.ConfigSearch{
		RefTable:    "products",
		NormFields:  []string{"code", "name"},
		Permissions: []string{"product.search"},
		RefDTO:      model.ProductSearchDTO{},
		CachePrefix: "product:search",
	})
}
