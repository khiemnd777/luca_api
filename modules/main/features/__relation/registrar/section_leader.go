package registrar

import (
	policy "github.com/khiemnd777/andy_api/modules/main/features/__relation/policy"
	"github.com/khiemnd777/andy_api/shared/logger"
)

func init() {
	logger.Debug("[RELATION] Register section - leader")
	policy.Register1("section_leader", policy.Config1{
		MainTable:    "sections",
		MainIDProp:   "ID",
		MainRefIDCol: "leader_id",

		RefTable:   "users",
		RefIDCol:   "id",
		RefNameCol: "name",
		RefFields:  []string{"id", "name"},

		Permissions: []string{"staff.view"},
		CachePrefix: "staff",
	})
	policy.RegisterRefSearch("section_leader", policy.ConfigSearch{
		RefTable:     "users",
		Alias:        "u",
		NormFields:   []string{"u.name"},
		RefFields:    []string{"id", "name"},
		SelectFields: []string{"u.id", "u.name"},
		Permissions:  []string{"staff.search"},
		CachePrefix:  "staff:search",
		ExtraJoins: func() string {
			return `
				JOIN staffs s ON s.user_staff = u.id
				JOIN staff_sections ss ON ss.staff_id = s.id 
				JOIN user_roles ur ON ur.user_id = u.id
				JOIN roles r ON r.id = ur.role_id
			`
		},
	})
}
