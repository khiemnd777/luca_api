package relation

type Config1 struct {
	MainTable  string // "orders"
	MainIDProp string // "ID"

	// foreign key
	MainRefIDCol   string  // "customer_id"
	MainRefNameCol *string // "customer_name"

	UpsertedIDProp   string  // "CustomerID"
	UpsertedNameProp *string // "CustomerName"

	RefTable   string // customers
	RefIDCol   string // "id"
	RefNameCol string // "name"
	RefDTO     any    // model.CustomerDTO{}

	Permissions []string
	CachePrefix string
}
