package relation

type Config1N struct {
	RefTable    string
	FKCol       string
	RefFields   []string
	Permissions []string
	CachePrefix string
}
