package relation

type Config1 struct {
	RefTable    string // customers
	IDCol       string // "id"
	RefDTO      any    // model.CustomerDTO{}
	Permissions []string
	CachePrefix string
}
