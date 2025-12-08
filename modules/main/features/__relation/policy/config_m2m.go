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

type ConfigM2M struct {
	// Schema
	MainTable string // ví dụ: "materials"
	RefTable  string // ví dụ: "suppliers"

	EntityPropMainID    string // e.g. "ID"
	DTOPropRefIDs       string // e.g. "SupplierIDs"
	DTOPropDisplayNames string // e.g. "SupplierNames"

	RefList *RefListConfig
}
