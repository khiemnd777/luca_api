package registrar

import (
	policy "github.com/khiemnd777/andy_api/modules/main/features/__relation/policy"
	"github.com/khiemnd777/andy_api/shared/logger"
	"github.com/khiemnd777/andy_api/shared/utils"
)

func init() {
	logger.Debug("[RELATION] Register order - clinic")
	policy.Register1("orders_clinics", policy.Config1{
		MainTable:      "orders",
		MainIDProp:     "ID",
		MainRefIDCol:   "clinic_id",
		MainRefNameCol: utils.Ptr("clinic_name"),

		RefTable:   "clinics",
		RefIDCol:   "id",
		RefNameCol: "name",
		RefFields:  []string{"id", "name"},

		UpsertedIDProp:   "ClinicID",
		UpsertedNameProp: utils.Ptr("ClinicName"),

		Permissions: []string{"clinic.view"},
		CachePrefix: "clinic",
	})
	policy.RegisterRefSearch("orders_clinics", policy.ConfigSearch{
		RefTable:    "clinics",
		NormFields:  []string{"name"},
		RefFields:   []string{"id", "name"},
		Permissions: []string{"clinic.search"},
		CachePrefix: "clinic:search",
	})
}
