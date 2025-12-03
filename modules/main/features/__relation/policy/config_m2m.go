package relation

type RefListConfig struct {
	Permissions []string
	RefDTO      any
	CachePrefix string
}

type RefSearchConfig struct {
	Permissions []string
	RefDTO      any
	CachePrefix string
}

type ConfigM2M struct {
	// Schema
	MainTable string // ví dụ: "materials"
	RefTable  string // ví dụ: "suppliers"

	GetMainID func(entity any) (int, error)
	GetIDs    func(input any) ([]int, error)
	SetResult func(output any, resIDs []int, resAsStr *string, res []string) error

	GetRefList   *RefListConfig
	GetRefSearch *RefSearchConfig
}
