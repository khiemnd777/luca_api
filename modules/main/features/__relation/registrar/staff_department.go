package registrar

import (
	policy "github.com/khiemnd777/andy_api/modules/main/features/__relation/policy"
	"github.com/khiemnd777/andy_api/shared/logger"
)

func init() {
	logger.Debug("[RELATION] Register staff - department")
	policy.RegisterRefSearch("staff_department", policy.ConfigSearch{
		RefTable:    "departments",
		Alias:       "d",
		NormFields:  []string{"d.name"},
		RefFields:   []string{"id", "name"},
		Permissions: []string{"department.view"},
		CachePrefix: "department:list",
	})
}
