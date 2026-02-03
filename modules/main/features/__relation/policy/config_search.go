package relation

type ConfigSearch struct {
	RefTable     string
	Alias        string
	NormFields   []string // []string{"code", "customer_name"}
	RefFields    []string
	SelectFields []string
	Permissions  []string
	CachePrefix  string
	ExtraJoins   func() string
	ExtraWhere   func(params ExtraWhereParams, args *[]any) string
	OrderRows    func([]map[string]any) []map[string]any
}

type ExtraWhereParams struct {
	DepartmentID int
}
