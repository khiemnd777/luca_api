package registrar

import (
	policy "github.com/khiemnd777/andy_api/modules/main/features/__relation/policy"
	"github.com/khiemnd777/andy_api/shared/logger"
)

func init() {
	logger.Debug("[RELATION] Register sections - processes")
	policy.RegisterM2M("sections_processes",
		policy.ConfigM2M{
			MainTable:           "sections",
			RefTable:            "processes",
			EntityPropMainID:    "ID",
			DTOPropRefIDs:       "ProcessIDs",
			DTOPropDisplayNames: "ProcessNames",

			RefList: &policy.RefListConfig{
				Permissions: []string{"process.view"},
				RefFields:   []string{"id", "code", "name"},
				CachePrefix: "process:list",
			},
		},
	)
	policy.RegisterRefSearch("sections_processes", policy.ConfigSearch{
		RefTable:    "processes",
		NormFields:  []string{"code", "name"},
		RefFields:   []string{"id", "code", "name"},
		Permissions: []string{"process.search"},
		CachePrefix: "process:list",
	})
}
