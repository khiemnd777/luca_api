package registrar

import (
	policy "github.com/khiemnd777/andy_api/modules/main/features/__relation/policy"
	"github.com/khiemnd777/andy_api/shared/logger"
	"github.com/khiemnd777/andy_api/shared/utils"
)

func init() {
	logger.Debug("[RELATION] Register order - customer")
	policy.Register1("orders_customers", policy.Config1{
		MainTable:      "orders",
		MainIDProp:     "ID",
		MainRefIDCol:   "customer_id",
		MainRefNameCol: utils.Ptr("customer_name"),

		RefTable:   "customers",
		RefIDCol:   "id",
		RefNameCol: "name",
		RefFields:  []string{"id", "name", "code"},

		UpsertedIDProp:   "CustomerID",
		UpsertedNameProp: utils.Ptr("CustomerName"),

		Permissions: []string{"customer.view"},
		CachePrefix: "customer",
	})
	policy.RegisterRefSearch("orders_customers", policy.ConfigSearch{
		RefTable:    "customers",
		NormFields:  []string{"code", "name"},
		RefFields:   []string{"id", "name", "code"},
		Permissions: []string{"customer.search"},
		CachePrefix: "customer:search",
	})
}
