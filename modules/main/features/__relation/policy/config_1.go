package relation

type Config1 struct {
	// Main
	MainTable      string  // "orders"
	MainIDProp     string  // "ID"
	MainRefIDCol   string  // "customer_id"
	MainRefNameCol *string // "customer_name"

	// Upsert
	UpsertedIDProp   string  // "CustomerID"
	UpsertedNameProp *string // "CustomerName"

	// Ref
	RefTable   string // customers
	RefIDCol   string // "id"
	RefFields  []string
	RefNameCol string // "name"

	// Get1
	Permissions []string
	CachePrefix string
}
