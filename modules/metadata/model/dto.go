package model

import "database/sql"

type Collection struct {
	ID   int    `json:"id"`
	Slug string `json:"slug"`
	Name string `json:"name"`
}

type Field struct {
	ID           int            `json:"id"`
	CollectionID int            `json:"collection_id"`
	Name         string         `json:"name"`
	Label        string         `json:"label"`
	Type         string         `json:"type"`
	Required     bool           `json:"required"`
	Unique       bool           `json:"unique"`
	DefaultValue sql.NullString `json:"default_value"`
	Options      sql.NullString `json:"options"`
	OrderIndex   int            `json:"order_index"`
	Visibility   string         `json:"visibility"`
	Relation     sql.NullString `json:"relation"`
}
