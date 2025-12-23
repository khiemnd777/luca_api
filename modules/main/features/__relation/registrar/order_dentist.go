package registrar

import (
	policy "github.com/khiemnd777/andy_api/modules/main/features/__relation/policy"
	"github.com/khiemnd777/andy_api/shared/logger"
	"github.com/khiemnd777/andy_api/shared/utils"
)

func init() {
	logger.Debug("[RELATION] Register order - dentist")
	policy.Register1("orders_dentists", policy.Config1{
		MainTable:      "orders",
		MainIDProp:     "ID",
		MainRefIDCol:   "dentist_id",
		MainRefNameCol: utils.Ptr("dentist_name"),

		RefTable:   "dentists",
		RefIDCol:   "id",
		RefNameCol: "name",
		RefFields:  []string{"id", "name"},

		UpsertedIDProp:   "DentistID",
		UpsertedNameProp: utils.Ptr("DentistName"),

		Permissions: []string{"clinic.view"},
		CachePrefix: "dentist",
	})
	policy.RegisterRefSearch("orders_dentists", policy.ConfigSearch{
		RefTable:    "dentists",
		NormFields:  []string{"name"},
		RefFields:   []string{"id", "name"},
		Permissions: []string{"clinic.search"},
		CachePrefix: "dentist:search",
	})
}
