package registrar

import (
	"fmt"

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
		ExtraWhere: func(params policy.ExtraWhereParams, args *[]any) string {
			*args = append(*args, params.DepartmentID)
			return fmt.Sprintf("r.deleted_at IS NULL AND r.department_id = $%d::INT", len(*args))
		},
		CachePrefix: "category_process:list",
	})
}
