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

	MainIDProp  string // e.g. "ID"
	RefIDsProp  string // e.g. "SupplierIDs"
	DisplayProp string // e.g. "SupplierNames"

	GetRefList   *RefListConfig
	GetRefSearch *RefSearchConfig
}
