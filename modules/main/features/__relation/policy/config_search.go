package relation

type ConfigSearch struct {
	RefTable    string
	NormFields  []string // []string{"code", "customer_name"}
	RefFields   []string
	Permissions []string
	CachePrefix string
}
