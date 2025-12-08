package relation

type RefListConfig struct {
	Permissions []string
	RefFields   []string
	CachePrefix string
}

type RefSearchConfig struct {
	Permissions []string
	RefFields   []string
	CachePrefix string
}

type ExtraM2MField struct {
	Column     string // column name in the m2m table, e.g. "color"
	EntityProp string // source field from the main entity, e.g. "Color"
}

// RefValueCacheColumn defines how to denormalize data from the m2m table into the ref table.
// Example: processes.section_name <- section_processes.section_name.
type RefValueCacheColumn struct {
	RefColumn string // column name in the ref table to update, e.g. "section_name"
	M2MColumn string // column name in the m2m table to read from, e.g. "section_name"
}

type RefValueCacheConfig struct {
	Columns []RefValueCacheColumn
}

type ConfigM2M struct {
	// Schema
	MainTable string // ví dụ: "materials"
	RefTable  string // ví dụ: "suppliers"

	EntityPropMainID    string // e.g. "ID"
	DTOPropRefIDs       string // e.g. "SupplierIDs"
	DTOPropDisplayNames string // e.g. "SupplierNames"

	RefList *RefListConfig
	// Optional extra columns to insert into the m2m table using values from the main entity.
	ExtraFields []ExtraM2MField
	// Optional: denormalize the referenced name into the m2m table, e.g. "process_name".
	RefNameColumn string
	// Optional: cache values from the m2m table back into the ref table.
	RefValueCache *RefValueCacheConfig
}
