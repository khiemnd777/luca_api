package customfields

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/khiemnd777/andy_api/shared/cache"
)

var (
	ErrUnknownField  = errors.New("unknown custom field")
	ErrRequired      = errors.New("required field missing")
	ErrInvalidType   = errors.New("invalid type")
	ErrInvalidOption = errors.New("invalid option")
)

type Store interface {
	GetIDBySlug(ctx context.Context, slug string) (*int, error)
	LoadSchema(ctx context.Context, collectionSlug string) (*Schema, error)
}

type PGStore struct{ DB *sql.DB }

func (s *PGStore) GetIDBySlug(ctx context.Context, slug string) (*int, error) {
	var collID int
	if err := s.DB.QueryRowContext(ctx, `SELECT id FROM collections WHERE slug=$1`, slug).Scan(&collID); err != nil {
		return nil, fmt.Errorf("load collection: %w", err)
	}
	return &collID, nil
}

func (s *PGStore) LoadSchema(ctx context.Context, slug string) (*Schema, error) {
	var collID int
	if err := s.DB.QueryRowContext(ctx, `SELECT id FROM collections WHERE slug=$1`, slug).Scan(&collID); err != nil {
		return nil, fmt.Errorf("load collection: %w", err)
	}
	rows, err := s.DB.QueryContext(ctx, `
        SELECT name, label, type, required, "unique", default_value, options, visibility
        FROM fields
        WHERE collection_id=$1
        ORDER BY order_index ASC, id ASC
    `, collID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var defs []FieldDef
	for rows.Next() {
		var f FieldDef
		var defJSON, optJSON []byte
		if err := rows.Scan(&f.Name, &f.Label, &f.Type, &f.Required, &f.Unique, &defJSON, &optJSON, &f.Visibility); err != nil {
			return nil, err
		}
		if len(defJSON) > 0 {
			_ = json.Unmarshal(defJSON, &f.DefaultValue)
		}
		if len(optJSON) > 0 {
			_ = json.Unmarshal(optJSON, &f.Options)
		}
		defs = append(defs, f)
	}
	return &Schema{Collection: slug, Fields: defs}, nil
}

type Manager struct {
	store Store
}

func NewManager(store Store) *Manager {
	return &Manager{store: store}
}

func (m *Manager) GetSchema(ctx context.Context, slug string) (*Schema, error) {
	collID, err := m.store.GetIDBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	return cache.Get(fmt.Sprintf("metadata:schema:i%d", collID), time.Hour*168, func() (*Schema, error) {
		return m.store.LoadSchema(ctx, slug)
	})
}
