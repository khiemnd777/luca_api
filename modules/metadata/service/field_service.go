package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/khiemnd777/andy_api/modules/metadata/model"
	"github.com/khiemnd777/andy_api/modules/metadata/repository"
	"github.com/khiemnd777/andy_api/shared/cache"
)

type FieldService struct {
	fields *repository.FieldRepository
	cols   *repository.CollectionRepository
}

func NewFieldService(f *repository.FieldRepository, c *repository.CollectionRepository) *FieldService {
	return &FieldService{fields: f, cols: c}
}

type FieldInput struct {
	CollectionID int             `json:"collection_id"`
	Name         string          `json:"name"`
	Label        string          `json:"label"`
	Type         string          `json:"type"`
	Required     bool            `json:"required"`
	Unique       bool            `json:"unique"`
	DefaultValue json.RawMessage `json:"default_value"`
	Options      json.RawMessage `json:"options"`
	OrderIndex   int             `json:"order_index"`
	Visibility   string          `json:"visibility"`
	Relation     json.RawMessage `json:"relation"`
}

// TTL
const (
	ttlFieldList = 2 * time.Minute
	ttlFieldItem = 10 * time.Minute
)

// ------- cache keys (phối hợp với CollectionService keys đã dùng trước đó) -------
func keyFieldsByCollection(collectionID int) string {
	return fmt.Sprintf("fields:collection:%d", collectionID)
}
func keyFieldByID(id int) string {
	return fmt.Sprintf("fields:id:%d", id)
}

// Trùng format với phần CollectionService đã cache trước đó:
func keyCollectionByID(id int, withFields bool) string {
	return fmt.Sprintf("collections:id:%d:f=%t", id, withFields)
}
func keyCollectionBySlug(slug string, withFields bool) string {
	return fmt.Sprintf("collections:slug:%s:f=%t", normalizeSlug(slug), withFields)
}

func (s *FieldService) ListByCollection(ctx context.Context, collectionID int) ([]model.Field, error) {
	key := keyFieldsByCollection(collectionID)

	type fieldList = []model.Field
	list, err := cache.Get(key, ttlFieldList, func() (*fieldList, error) {
		items, err := s.fields.ListByCollectionID(ctx, collectionID)
		if err != nil {
			return nil, err
		}
		l := fieldList(items)
		return &l, nil
	})
	if err != nil {
		return nil, err
	}
	return *list, nil
}

func (s *FieldService) Get(ctx context.Context, id int) (*model.Field, error) {
	key := keyFieldByID(id)
	return cache.Get(key, ttlFieldItem, func() (*model.Field, error) {
		return s.fields.Get(ctx, id)
	})
}

func (s *FieldService) Create(ctx context.Context, in FieldInput) (*model.Field, error) {
	if _, err := s.cols.GetByID(ctx, in.CollectionID, false); err != nil {
		return nil, fmt.Errorf("collection not found")
	}
	in.Name = strings.TrimSpace(in.Name)
	in.Label = strings.TrimSpace(in.Label)
	if in.Name == "" || in.Label == "" {
		return nil, fmt.Errorf("name/label required")
	}

	f := &model.Field{
		CollectionID: in.CollectionID,
		Name:         in.Name,
		Label:        in.Label,
		Type:         in.Type,
		Required:     in.Required,
		Unique:       in.Unique,
		DefaultValue: toNullString(in.DefaultValue),
		Options:      toNullString(in.Options),
		OrderIndex:   in.OrderIndex,
		Visibility:   firstOrDefault(strings.TrimSpace(in.Visibility), "public"),
		Relation:     toNullString(in.Relation),
	}

	created, err := s.fields.Create(ctx, f)
	if err != nil {
		return nil, err
	}

	cache.InvalidateKeys(
		keyFieldsByCollection(in.CollectionID),
		keyCollectionByID(in.CollectionID, true),
	)
	cache.InvalidateKeys("collections:slug:*")

	return created, nil
}

func (s *FieldService) Update(ctx context.Context, id int, in FieldInput) (*model.Field, error) {
	cur, err := s.fields.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	oldColID := cur.CollectionID

	if in.CollectionID != 0 && in.CollectionID != cur.CollectionID {
		if _, err := s.cols.GetByID(ctx, in.CollectionID, false); err != nil {
			return nil, fmt.Errorf("collection not found")
		}
		cur.CollectionID = in.CollectionID
	}
	if strings.TrimSpace(in.Name) != "" {
		cur.Name = strings.TrimSpace(in.Name)
	}
	if strings.TrimSpace(in.Label) != "" {
		cur.Label = strings.TrimSpace(in.Label)
	}
	if strings.TrimSpace(in.Type) != "" {
		cur.Type = strings.TrimSpace(in.Type)
	}
	cur.Required = in.Required
	cur.Unique = in.Unique
	if len(in.DefaultValue) > 0 {
		cur.DefaultValue = toNullString(in.DefaultValue)
	}
	if len(in.Options) > 0 {
		cur.Options = toNullString(in.Options)
	}
	if in.OrderIndex != 0 {
		cur.OrderIndex = in.OrderIndex
	}
	if strings.TrimSpace(in.Visibility) != "" {
		cur.Visibility = strings.TrimSpace(in.Visibility)
	}
	if len(in.Relation) > 0 {
		cur.Relation = toNullString(in.Relation)
	}

	updated, err := s.fields.Update(ctx, cur)
	if err != nil {
		return nil, err
	}

	keys := []string{
		keyFieldByID(id),
		keyFieldsByCollection(oldColID),
		keyCollectionByID(oldColID, true),
	}
	if cur.CollectionID != oldColID {
		keys = append(keys,
			keyFieldsByCollection(cur.CollectionID),
			keyCollectionByID(cur.CollectionID, true),
		)
	}
	cache.InvalidateKeys(keys...)
	cache.InvalidateKeys("collections:slug:*")

	return updated, nil
}

func (s *FieldService) Delete(ctx context.Context, id int) error {
	cur, err := s.fields.Get(ctx, id)
	if err != nil {
		return err
	}
	if err := s.fields.Delete(ctx, id); err != nil {
		return err
	}

	cache.InvalidateKeys(
		keyFieldByID(id),
		keyFieldsByCollection(cur.CollectionID),
		keyCollectionByID(cur.CollectionID, true),
	)
	cache.InvalidateKeys("collections:slug:*")

	return nil
}

func toNullString(b json.RawMessage) sql.NullString {
	if len(b) == 0 {
		return sql.NullString{}
	}
	return sql.NullString{String: string(b), Valid: true}
}

func firstOrDefault(v, def string) string {
	if v == "" {
		return def
	}
	return v
}
