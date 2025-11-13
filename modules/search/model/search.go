package model

import "time"

type Row struct {
	EntityType string         `json:"entity_type"`
	EntityID   int64          `json:"entity_id"`
	Title      string         `json:"title"`
	Subtitle   *string        `json:"subtitle,omitempty"`
	Keywords   *string        `json:"keywords,omitempty"`
	Attributes map[string]any `json:"attributes,omitempty"`
	Rank       *float64       `json:"rank,omitempty"` // ts_rank từ lớp full-text
	UpdatedAt  time.Time      `json:"updated_at"`
}

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
