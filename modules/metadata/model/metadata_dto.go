package model

import (
	"database/sql"
	"encoding/json"
)

type Collection struct {
	ID          int             `json:"id"`
	Slug        string          `json:"slug"`
	Name        string          `json:"name"`
	ShowIf      *sql.NullString `json:"show_if,omitempty"`
	Integration bool            `json:"integration,omitempty"`
	Group       *string         `json:"group,omitempty"`
}

type CollectionDTO struct {
	ID          int     `json:"id"`
	Slug        string  `json:"slug"`
	Name        string  `json:"name"`
	ShowIf      *string `json:"show_if,omitempty"`
	Integration bool    `json:"integration,omitempty"`
	Group       *string `json:"group,omitempty"`
}

type Field struct {
	ID             int             `json:"id"`
	CollectionID   int             `json:"collection_id"`
	CollectionSlug string          `json:"collection_slug"`
	Name           string          `json:"name"`
	Label          string          `json:"label"`
	Type           string          `json:"type"`
	Required       bool            `json:"required"`
	Unique         bool            `json:"unique"`
	Table          bool            `json:"table"`
	Form           bool            `json:"form"`
	Search         bool            `json:"search"`
	DefaultValue   *sql.NullString `json:"default_value"`
	Options        *sql.NullString `json:"options"`
	OrderIndex     int             `json:"order_index"`
	Visibility     string          `json:"visibility"`
	Relation       *sql.NullString `json:"relation"`
}

type FieldDTO struct {
	ID             int     `json:"id"`
	CollectionID   int     `json:"collection_id"`
	CollectionSlug string  `json:"collection_slug"`
	Name           string  `json:"name"`
	Label          string  `json:"label"`
	Type           string  `json:"type"`
	Required       bool    `json:"required"`
	Unique         bool    `json:"unique"`
	Table          bool    `json:"table"`
	Form           bool    `json:"form"`
	Search         bool    `json:"search"`
	DefaultValue   *string `json:"default_value"`
	Options        *string `json:"options"`
	OrderIndex     int     `json:"order_index"`
	Visibility     string  `json:"visibility"`
	Relation       *string `json:"relation"`
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
	Search       bool             `json:"search"`
	DefaultValue *json.RawMessage `json:"default_value"`
	Options      *json.RawMessage `json:"options"`
	OrderIndex   int              `json:"order_index"`
	Visibility   string           `json:"visibility"`
	Relation     *json.RawMessage `json:"relation"`
}
