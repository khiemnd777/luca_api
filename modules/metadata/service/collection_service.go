package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/khiemnd777/andy_api/modules/metadata/model"
	"github.com/khiemnd777/andy_api/modules/metadata/repository"
	"github.com/khiemnd777/andy_api/shared/cache"
)

type CollectionService struct {
	repo *repository.CollectionRepository
}

func NewCollectionService(r *repository.CollectionRepository) *CollectionService {
	return &CollectionService{repo: r}
}

const (
	ttlCollectionList = 30 * time.Second
	ttlCollectionItem = 5 * time.Minute
)

func cacheKeyList(query string, limit, offset int, withFields bool, tag *string, table, form bool) string {
	tagVal := "nil"
	if tag != nil {
		tagVal = *tag
	}

	return fmt.Sprintf(
		"collections:list:q=%s:l=%d:o=%d:f=%t:t=%s:tb=%t:fm=%t",
		query, limit, offset, withFields, tagVal, table, form,
	)
}

func cacheKeyListIntegration(group, query string, limit, offset int, withFields bool, tag *string, table, form bool) string {
	tagVal := "nil"
	if tag != nil {
		tagVal = *tag
	}
	return fmt.Sprintf("collections:list:g=%s:q=%s:l=%d:o=%d:f=%t:t=%s:tb:%t:fm:%t", group, query, limit, offset, withFields, tagVal, table, form)
}

func cacheKeySlug(slug string, withFields bool, tag *string, table, form bool) string {
	tagVal := "nil"
	if tag != nil {
		tagVal = *tag
	}
	return fmt.Sprintf("collections:slug:%s:f=%t:t=%s:tb:%t:fm:%t", slug, withFields, tagVal, table, form)
}

func cacheKeyAvailableSlug(slug string, withFields bool, tag *string, table, form bool) string {
	tagVal := "nil"
	if tag != nil {
		tagVal = *tag
	}
	return fmt.Sprintf("collections:slug:%s:abl:f=%t:t=%s:tb:%t:fm:%t", slug, withFields, tagVal, table, form)
}

func cacheKeySlugAll(slug string) string {
	return fmt.Sprintf("collections:slug:%s:*", slug)
}

func cacheKeyID(id int, withFields bool, tag *string, table, form bool) string {
	tagVal := "nil"
	if tag != nil {
		tagVal = *tag
	}
	return fmt.Sprintf("collections:id:%d:f=%t:t=%s:tb:%t:fm:%t", id, withFields, tagVal, table, form)
}

func cacheKeyAvaialbleID(id int, withFields bool, tag *string, table, form bool) string {
	tagVal := "nil"
	if tag != nil {
		tagVal = *tag
	}
	return fmt.Sprintf("collections:id:%d:abl:f=%t:t=%s:tb:%t:fm:%t", id, withFields, tagVal, table, form)
}

func cacheKeyIDAll(id int) string {
	return fmt.Sprintf("collections:id:%d:*", id)
}

type ListCollectionsInput struct {
	Query      string
	Limit      int
	Offset     int
	WithFields bool
	Tag        *string
	Table      bool
	Form       bool
}

type CreateCollectionInput struct {
	Slug        string           `json:"slug"`
	Name        string           `json:"name"`
	ShowIf      *json.RawMessage `json:"show_if"`
	Integration bool             `json:"integration"`
	Group       *string          `json:"group"`
}

type UpdateCollectionInput struct {
	Slug        *string          `json:"slug"`
	Name        *string          `json:"name"`
	ShowIf      *json.RawMessage `json:"show_if"`
	Integration bool             `json:"integration"`
	Group       *string          `json:"group"`
}

var slugRegex = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

func normalizeSlug(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	s = strings.ReplaceAll(s, "_", "-")
	s = strings.ReplaceAll(s, " ", "-")
	return strings.Trim(s, "-")
}

func (s *CollectionService) ListIntegration(ctx context.Context, group, query string, limit, offset int, withFields bool, tag *string, table, form bool) ([]repository.CollectionWithFields, int, error) {
	key := cacheKeyListIntegration(group, query, limit, offset, withFields, tag, table, form)

	type result struct {
		Items []repository.CollectionWithFields
		Total int
	}

	r, err := cache.Get(key, ttlCollectionList, func() (*result, error) {
		items, total, err := s.repo.ListIntegration(ctx, group, query, limit, offset, withFields, tag, table, form)
		if err != nil {
			return nil, err
		}
		return &result{Items: items, Total: total}, nil
	})
	if err != nil {
		return nil, 0, err
	}

	return r.Items, r.Total, nil
}

func (s *CollectionService) List(ctx context.Context, in ListCollectionsInput) ([]repository.CollectionWithFields, int, error) {
	key := cacheKeyList(in.Query, in.Limit, in.Offset, in.WithFields, in.Tag, in.Table, in.Form)

	type result struct {
		Items []repository.CollectionWithFields
		Total int
	}

	r, err := cache.Get(key, ttlCollectionList, func() (*result, error) {
		items, total, err := s.repo.List(ctx, in.Query, in.Limit, in.Offset, in.WithFields, in.Tag, in.Table, in.Form)
		if err != nil {
			return nil, err
		}
		return &result{Items: items, Total: total}, nil
	})
	if err != nil {
		return nil, 0, err
	}
	return r.Items, r.Total, nil
}

func (s *CollectionService) GetBySlug(ctx context.Context, slug string, withFields bool, tag *string, table, form bool) (*repository.CollectionWithFields, error) {
	slug = normalizeSlug(slug)
	key := cacheKeySlug(slug, withFields, tag, table, form)

	return cache.Get(key, ttlCollectionItem, func() (*repository.CollectionWithFields, error) {
		return s.repo.GetBySlug(ctx, slug, withFields, tag, table, form, true, nil)
	})
}

func (s *CollectionService) GetByAvailableSlug(ctx context.Context, slug string, withFields bool, tag *string, table, form bool, entityData *map[string]any) (*repository.CollectionWithFields, error) {
	slug = normalizeSlug(slug)
	// key := cacheKeyAvailableSlug(slug, withFields, tag, table, form)

	// return cache.Get(key, ttlCollectionItem, func() (*repository.CollectionWithFields, error) {
	return s.repo.GetBySlug(ctx, slug, withFields, tag, table, form, false, entityData)
	// })
}

func (s *CollectionService) GetByID(ctx context.Context, id int, withFields bool, tag *string, table, form bool) (*repository.CollectionWithFields, error) {
	key := cacheKeyID(id, withFields, tag, table, form)

	return cache.Get(key, ttlCollectionItem, func() (*repository.CollectionWithFields, error) {
		return s.repo.GetByID(ctx, id, withFields, tag, table, form, true, nil)
	})
}

func (s *CollectionService) GetAvailableByID(ctx context.Context, id int, withFields bool, tag *string, table, form bool, entityData *map[string]any) (*repository.CollectionWithFields, error) {
	key := cacheKeyAvaialbleID(id, withFields, tag, table, form)

	return cache.Get(key, ttlCollectionItem, func() (*repository.CollectionWithFields, error) {
		return s.repo.GetByID(ctx, id, withFields, tag, table, form, false, entityData)
	})
}

func (s *CollectionService) Create(ctx context.Context, in CreateCollectionInput) (*model.CollectionDTO, error) {
	in.Slug = normalizeSlug(in.Slug)
	if !slugRegex.MatchString(in.Slug) {
		return nil, ErrBadSlug
	}
	if strings.TrimSpace(in.Name) == "" {
		return nil, ErrBadName
	}
	exists, err := s.repo.SlugExists(ctx, in.Slug, nil)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrConflict("slug already exists")
	}

	var showIfVal *sql.NullString = nil
	if in.ShowIf != nil && len(*in.ShowIf) > 0 {
		sif := toNullString(*in.ShowIf)
		showIfVal = &sif
	}

	c, err := s.repo.Create(ctx, normalizeSlug(in.Slug), in.Name, showIfVal, in.Integration, in.Group)
	if err != nil {
		return nil, err
	}
	cache.InvalidateKeys("collections:list:*")
	return c, nil

}

func (s *CollectionService) Update(ctx context.Context, id int, in UpdateCollectionInput) (*model.CollectionDTO, error) {
	var ex *int = &id
	if in.Slug != nil {
		slug := normalizeSlug(*in.Slug)
		if !slugRegex.MatchString(slug) {
			return nil, ErrBadSlug
		}
		ok, err := s.repo.SlugExists(ctx, slug, ex)
		if err != nil {
			return nil, err
		}
		if ok {
			return nil, ErrConflict("slug already exists")
		}
		in.Slug = &slug
	}
	if in.Name != nil {
		name := strings.TrimSpace(*in.Name)
		if name == "" {
			return nil, ErrBadName
		}
		in.Name = &name
	}
	var showIfVal *sql.NullString = nil
	if in.ShowIf != nil && len(*in.ShowIf) > 0 {
		sif := toNullString(*in.ShowIf)
		showIfVal = &sif
	}

	c, err := s.repo.Update(ctx, id, in.Slug, in.Name, showIfVal, &in.Integration, in.Group)
	if err != nil {
		return nil, err
	}

	cache.InvalidateKeys(cacheKeyIDAll(id), fmt.Sprintf("metadata:schema:i%d", id))
	if in.Slug != nil {
		cache.InvalidateKeys(
			cacheKeySlugAll(*in.Slug),
		)
	}
	cache.InvalidateKeys("collections:list:*")

	return c, nil
}

func (s *CollectionService) Delete(ctx context.Context, id int) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	cache.InvalidateKeys(
		cacheKeyIDAll(id),
		fmt.Sprintf("metadata:schema:i%d", id),
	)
	cache.InvalidateKeys("collections:list:*", "collections:slug:*")
	return nil
}

// errors
type ErrConflict string

func (e ErrConflict) Error() string { return string(e) }

var (
	ErrBadSlug = simpleErr("invalid slug (lowercase letters, numbers and dashes)")
	ErrBadName = simpleErr("name must not be empty")
)

type simpleErr string

func (e simpleErr) Error() string { return string(e) }
