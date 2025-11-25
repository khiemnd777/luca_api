package registrar

import (
	model "github.com/khiemnd777/andy_api/modules/main/features/__model"
	policy "github.com/khiemnd777/andy_api/modules/main/features/__relation/policy"
	"github.com/khiemnd777/andy_api/shared/logger"
	"github.com/khiemnd777/andy_api/shared/utils"
)

func init() {
	logger.Debug("[RELATION] Register order - customer")
	policy.Register1("order", policy.Config1{
		MainTable:      "orders",
		MainIDProp:     "ID",
		MainRefIDCol:   "customer_id",
		MainRefNameCol: utils.Ptr("customer_name"),

		RefTable:   "customers",
		RefIDCol:   "id",
		RefNameCol: "name",
		RefDTO:     model.CustomerDTO{},

		UpsertedIDProp:   "CustomerID",
		UpsertedNameProp: utils.Ptr("CustomerName"),

		Permissions: []string{"customer.view"},
		CachePrefix: "customer",
	})
	policy.RegisterRefSearch("order", policy.ConfigSearch{
		RefTable:    "customers",
		NormFields:  []string{"code", "name"},
		Permissions: []string{"customer.search"},
		RefDTO:      model.CustomerDTO{},
		CachePrefix: "customer:search",
	})
}
