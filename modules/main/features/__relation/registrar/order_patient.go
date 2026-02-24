package registrar

import (
	policy "github.com/khiemnd777/andy_api/modules/main/features/__relation/policy"
	"github.com/khiemnd777/andy_api/shared/logger"
	"github.com/khiemnd777/andy_api/shared/utils"
)

func init() {
	logger.Debug("[RELATION] Register order - patient")
	policy.Register1("orders_patients", policy.Config1{
		MainTable:      "orders",
		MainIDProp:     "ID",
		MainRefIDCol:   "patient_id",
		MainRefNameCol: utils.Ptr("patient_name"),

		RefTable:   "patients",
		RefIDCol:   "id",
		RefNameCol: "name",
		RefFields:  []string{"id", "name"},

		UpsertedIDProp:   "PatientID",
		UpsertedNameProp: utils.Ptr("PatientName"),

		Permissions: []string{"clinic.view"},
		CachePrefix: "patient",
	})
	policy.RegisterRefSearch("orders_patients", policy.ConfigSearch{
		RefTable:     "clinic_patients",
		Alias:        "cp",
		NormFields:   []string{"r.name"},
		RefFields:    []string{"id", "name"},
		SelectFields: []string{"r.id", "r.name"},
		Permissions:  []string{"clinic.search"},
		CachePrefix:  "patient:search",
		ExtraJoins: func() string {
			return `
				JOIN patients r ON r.id = cp.patient_id
			`
		},
	})
}
