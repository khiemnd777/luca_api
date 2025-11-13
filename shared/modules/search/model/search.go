package model

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
