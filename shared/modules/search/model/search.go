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

type UnlinkDoc struct {
	EntityType string `json:"entity_type"`
	EntityID   int64  `json:"entity_id"`
}

type Doc struct {
	EntityType string         `json:"entity_type"`
	EntityID   int64          `json:"entity_id"`
	Title      string         `json:"title"`
	Subtitle   *string        `json:"subtitle,omitempty"`
	Keywords   *string        `json:"keywords,omitempty"`
	Content    *string        `json:"content,omitempty"`
	Attributes map[string]any `json:"attributes,omitempty"`
	OrgID      *int64         `json:"org_id,omitempty"`
	OwnerID    *int64         `json:"owner_id,omitempty"`
	ACLHash    *string        `json:"acl_hash,omitempty"`
}
