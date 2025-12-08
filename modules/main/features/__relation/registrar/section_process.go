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
			ExtraFields: []policy.ExtraM2MField{
				{Column: "color", EntityProp: "Color"},
				{Column: "section_name", EntityProp: "Name"},
			},
			RefNameColumn: "process_name",
			RefValueCache: &policy.RefValueCacheConfig{
				Columns: []policy.RefValueCacheColumn{
					{RefColumn: "section_id", M2MColumn: "section_id"},
					{RefColumn: "section_name", M2MColumn: "section_name"},
					{RefColumn: "color", M2MColumn: "color"},
				},
			},

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
		ExtraWhere: func(args *[]any) string {
			return `
				r.deleted_at IS NULL AND
				NOT EXISTS (
					SELECT 1 FROM section_processes sp
					WHERE sp.process_id = r.id
				)
			`
		},
	})
}
