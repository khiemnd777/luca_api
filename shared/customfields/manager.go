package customfields

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	ErrUnknownField  = errors.New("unknown custom field")
	ErrRequired      = errors.New("required field missing")
	ErrInvalidType   = errors.New("invalid type")
	ErrInvalidOption = errors.New("invalid option")
)

type Store interface {
	LoadSchema(ctx context.Context, collectionSlug string) (*Schema, error)
}

type PGStore struct{ DB *sql.DB }

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

type cacheEntry struct {
	s   *Schema
	exp time.Time
}
type Cache struct {
	mu  sync.RWMutex
	ttl time.Duration
	m   map[string]cacheEntry
}

func NewCache(ttl time.Duration) *Cache { return &Cache{ttl: ttl, m: map[string]cacheEntry{}} }
func (c *Cache) Get(key string) (*Schema, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	it, ok := c.m[key]
	if !ok || time.Now().After(it.exp) {
		return nil, false
	}
	return it.s, true
}
func (c *Cache) Set(key string, s *Schema) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.m[key] = cacheEntry{s: s, exp: time.Now().Add(c.ttl)}
}

type Manager struct {
	store Store
	cache *Cache
}

func NewManager(store Store, cacheTTL time.Duration) *Manager {
	return &Manager{store: store, cache: NewCache(cacheTTL)}
}

func (m *Manager) GetSchema(ctx context.Context, slug string) (*Schema, error) {
	if s, ok := m.cache.Get(slug); ok {
		return s, nil
	}
	s, err := m.store.LoadSchema(ctx, slug)
	if err != nil {
		return nil, err
	}
	m.cache.Set(slug, s)
	return s, nil
}
