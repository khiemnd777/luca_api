package registrar

import (
	"fmt"

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
		ExtraWhere: func(params policy.ExtraWhereParams, args *[]any) string {
			*args = append(*args, params.DepartmentID)
			return fmt.Sprintf("r.department_id = $%d::INT", len(*args))
		},
		CachePrefix: "product:search",
	})
}
