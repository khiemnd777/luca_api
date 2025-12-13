package registrar

import (
	policy "github.com/khiemnd777/andy_api/modules/main/features/__relation/policy"
	"github.com/khiemnd777/andy_api/shared/logger"
)

func init() {
	logger.Debug("[RELATION] Register categories - processes")
	policy.RegisterM2M("categories_processes",
		policy.ConfigM2M{
			MainTable:        "categories",
			RefTable:         "processes",
			EntityPropMainID: "ID",
			DTOPropRefIDs:    "ProcessIDs",

			RefList: &policy.RefListConfig{
				Permissions: []string{"process.view"},
				RefFields:   []string{"id", "code", "name", "section_name", "color"},
				CachePrefix: "category_process:list",
			},
		},
	)
	policy.RegisterRefSearch("categories_processes", policy.ConfigSearch{
		RefTable:    "processes",
		NormFields:  []string{"code", "name"},
		RefFields:   []string{"id", "code", "name", "section_name", "color"},
		Permissions: []string{"process.search"},
		CachePrefix: "category_process:list",
	})
}
