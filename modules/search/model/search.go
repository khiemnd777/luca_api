package model

type Options struct {
	Query           string            // raw keyword (có dấu). SQL sẽ unaccent.
	Types           []string          // filter theo loại. Nil/empty = all.
	OrgID           *int64            // scope theo org (nếu có)
	OwnerID         *int64            // scope theo owner (optional)
	Filters         map[string]string // attributes filters: key -> exact value (mở rộng sau)
	Limit           int
	Offset          int
	UseTrgmFallback bool // fallback nếu ít kết quả full-text
}
