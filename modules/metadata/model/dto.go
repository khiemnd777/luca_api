package model

import (
	"database/sql"
	"encoding/json"
)

type Collection struct {
	ID   int    `json:"id"`
	Slug string `json:"slug"`
	Name string `json:"name"`
}

type Field struct {
	ID           int             `json:"id"`
	CollectionID int             `json:"collection_id"`
	Name         string          `json:"name"`
	Label        string          `json:"label"`
	Type         string          `json:"type"`
	Required     bool            `json:"required"`
	Unique       bool            `json:"unique"`
	Table        bool            `json:"table"`
	Form         bool            `json:"form"`
	DefaultValue *sql.NullString `json:"default_value"`
	Options      *sql.NullString `json:"options"`
	OrderIndex   int             `json:"order_index"`
	Visibility   string          `json:"visibility"`
	Relation     *sql.NullString `json:"relation"`
}

type FieldInput struct {
	CollectionID int              `json:"collection_id"`
	Name         string           `json:"name"`
	Label        string           `json:"label"`
	Type         string           `json:"type"`
	Required     bool             `json:"required"`
	Unique       bool             `json:"unique"`
	Table        bool             `json:"table"`
	Form         bool             `json:"form"`
	DefaultValue *json.RawMessage `json:"default_value"`
	Options      *json.RawMessage `json:"options"`
	OrderIndex   int              `json:"order_index"`
	Visibility   string           `json:"visibility"`
	Relation     *json.RawMessage `json:"relation"`
}
