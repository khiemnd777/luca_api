package registrar

import (
	policy "github.com/khiemnd777/andy_api/modules/main/features/__relation/policy"
	"github.com/khiemnd777/andy_api/shared/logger"
)

func __init() {
	logger.Debug("[RELATION] Register section - processes")
	policy.Register1N("sections_processes", policy.Config1N{
		// List
		RefTable:    "processes",
		IDCol:       "id",
		FKCol:       "section_id",
		RefFields:   []string{"id", "code", "name"},
		Permissions: []string{"process.view"},
		CachePrefix: "section_process",

		// Upsert
		IDProp:       "ID",
		ParentIDProp: "SectionID",
		InsertCols:   []string{"code", "name", "custom_fields"},
		InsertProps:  []string{"Code", "Name", "CustomFields"},
		ReturnCols:   []string{"id", "section_id", "section_name", "code", "name", "custom_fields", "created_at", "updated_at"},
		ReturnProps:  []string{"ID", "SectionID", "SectionName", "Code", "Name", "CustomFields", "CreatedAt", "UpdatedAt"},
		UpdatedAtCol: "updated_at",
	})
}
