package relation

type ConfigSearch struct {
	RefTable    string
	NormFields  []string // []string{"code", "customer_name"}
	Permissions []string
	RefDTO      any
	CachePrefix string
}
