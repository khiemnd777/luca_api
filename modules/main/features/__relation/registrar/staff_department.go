package registrar

import (
	policy "github.com/khiemnd777/andy_api/modules/main/features/__relation/policy"
	"github.com/khiemnd777/andy_api/shared/logger"
)

func init() {
	logger.Debug("[RELATION] Register staff - department")
	policy.RegisterRefSearch("staff_department", policy.ConfigSearch{
		RefTable:     "users",
		Alias:        "u",
		NormFields:   []string{"u.name"},
		RefFields:    []string{"id", "name"},
		SelectFields: []string{"u.id", "u.name"},
		ExtraJoins: func() string {
			return `
				JOIN staffs s ON s.user_staff = u.id
			`
		},
		Permissions: []string{"staff.search"},
		CachePrefix: "staff:search",
	})
}
