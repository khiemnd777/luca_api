package relation

type Config1N struct {
	RefTable    string
	FKCol       string
	RefDTO      any
	Permissions []string
	CachePrefix string
}
