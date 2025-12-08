package relation

type Config1N struct {
	RefTable    string
	IDCol       string
	FKCol       string
	RefFields   []string
	Permissions []string
	CachePrefix string

	IDProp       string
	ParentIDProp string
	InsertCols   []string
	InsertProps  []string
	ReturnCols   []string
	ReturnProps  []string
	UpdatedAtCol string
}
