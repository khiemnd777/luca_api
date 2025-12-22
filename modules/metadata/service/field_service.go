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

func keyCollectionByID(id int, withFields bool) string {
	return fmt.Sprintf("collections:id:%d:f=%t", id, withFields)
}

func (s *FieldService) ListByCollection(ctx context.Context, collectionID int) ([]*model.FieldDTO, error) {
	key := keyFieldsByCollection(collectionID)

	type fieldList = []*model.FieldDTO
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

func (s *FieldService) Get(ctx context.Context, id int) (*model.FieldDTO, error) {
	key := keyFieldByID(id)
	return cache.Get(key, ttlFieldItem, func() (*model.FieldDTO, error) {
		return s.fields.Get(ctx, id)
	})
}

func (s *FieldService) Create(ctx context.Context, in model.FieldInput) (*model.FieldDTO, error) {
	if _, err := s.cols.GetByID(ctx, in.CollectionID, false, nil, false, false, true, nil); err != nil {
		return nil, fmt.Errorf("collection not found")
	}
	in.Name = strings.TrimSpace(in.Name)
	in.Label = strings.TrimSpace(in.Label)
	if in.Name == "" || in.Label == "" {
		return nil, fmt.Errorf("name/label required")
	}

	var df *sql.NullString = nil
	if in.DefaultValue != nil && len(*in.DefaultValue) > 0 {
		ns := toNullString(*in.DefaultValue)
		df = &ns
	}

	var opt *sql.NullString = nil
	if in.Options != nil && len(*in.Options) > 0 {
		ns := toNullString(*in.Options)
		opt = &ns
	}

	var rel *sql.NullString = nil
	if in.Relation != nil && len(*in.Relation) > 0 {
		ns := toNullString(*in.Relation)
		rel = &ns
	}

	f := &model.Field{
		CollectionID: in.CollectionID,
		Name:         in.Name,
		Label:        in.Label,
		Type:         in.Type,
		Required:     in.Required,
		Unique:       in.Unique,
		Tag:          in.Tag,
		Table:        in.Table,
		Form:         in.Form,
		Search:       in.Search,
		DefaultValue: df,
		Options:      opt,
		OrderIndex:   in.OrderIndex,
		Visibility:   firstOrDefault(strings.TrimSpace(in.Visibility), "public"),
		Relation:     rel,
	}

	created, err := s.fields.Create(ctx, f)
	if err != nil {
		return nil, err
	}

	col, err := s.cols.GetByID(ctx, in.CollectionID, false, nil, false, false, false, nil)

	if err != nil {
		return nil, err
	}

	created.CollectionSlug = col.Slug

	cache.InvalidateKeys(
		keyFieldsByCollection(in.CollectionID),
		fmt.Sprintf("metadata:schema:i%d", in.CollectionID),
		keyCollectionByID(in.CollectionID, true),
	)
	cache.InvalidateKeys("collections:slug:*")

	return created, nil
}

func (s *FieldService) Update(ctx context.Context, id int, in model.FieldInput) (*model.FieldDTO, error) {
	cur, err := s.fields.GetRaw(ctx, id)
	if err != nil {
		return nil, err
	}

	oldColID := cur.CollectionID

	if in.CollectionID != 0 && in.CollectionID != cur.CollectionID {
		if _, err := s.cols.GetByID(ctx, in.CollectionID, false, nil, false, false, true, nil); err != nil {
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
	cur.Tag = in.Tag
	cur.Table = in.Table
	cur.Form = in.Form
	cur.Search = in.Search

	if in.DefaultValue != nil && len(*in.DefaultValue) > 0 {
		ns := toNullString(*in.DefaultValue)
		cur.DefaultValue = &ns
	} else {
		cur.DefaultValue = nil
	}

	if in.Options != nil && len(*in.Options) > 0 {
		ns := toNullString(*in.Options)
		cur.Options = &ns
	} else {
		cur.DefaultValue = nil
	}

	if in.OrderIndex != 0 {
		cur.OrderIndex = in.OrderIndex
	}
	if strings.TrimSpace(in.Visibility) != "" {
		cur.Visibility = strings.TrimSpace(in.Visibility)
	}
	if in.Relation != nil && len(*in.Relation) > 0 {
		ns := toNullString(*in.Relation)
		cur.Relation = &ns
	} else {
		cur.Relation = nil
	}

	updated, err := s.fields.Update(ctx, cur)
	if err != nil {
		return nil, err
	}

	col, err := s.cols.GetByID(ctx, in.CollectionID, false, nil, false, false, false, nil)

	if err != nil {
		return nil, err
	}

	updated.CollectionSlug = col.Slug

	keys := []string{
		keyFieldByID(id),
		keyFieldsByCollection(oldColID),
		keyCollectionByID(oldColID, true),
		fmt.Sprintf("metadata:schema:i%d", oldColID),
	}
	if cur.CollectionID != oldColID {
		keys = append(keys,
			keyFieldsByCollection(cur.CollectionID),
			keyCollectionByID(cur.CollectionID, true),
			fmt.Sprintf("metadata:schema:i%d", cur.CollectionID),
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
		fmt.Sprintf("metadata:schema:i%d", cur.CollectionID),
	)
	cache.InvalidateKeys("collections:slug:*")

	return nil
}

func (s *FieldService) Sort(ctx context.Context, collectionID int, ids []int) (*string, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	col, err := s.cols.GetByID(ctx, collectionID, false, nil, false, false, true, nil)

	if err != nil {
		return nil, fmt.Errorf("collection not found")
	}

	if err := s.fields.Sort(ctx, ids); err != nil {
		return nil, err
	}

	cache.InvalidateKeys(
		keyFieldsByCollection(collectionID),
		keyCollectionByID(collectionID, true),
		fmt.Sprintf("metadata:schema:i%d", collectionID),
	)
	cache.InvalidateKeys("collections:slug:*")

	return &col.Slug, nil
}

func toNullString(b json.RawMessage) sql.NullString {
	s := strings.TrimSpace(string(b))
	if s == "" || s == "null" || s == "\"\"" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

func firstOrDefault(v, def string) string {
	if v == "" {
		return def
	}
	return v
}
