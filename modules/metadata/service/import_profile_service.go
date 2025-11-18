package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/khiemnd777/andy_api/modules/metadata/model"
	"github.com/khiemnd777/andy_api/modules/metadata/repository"
)

type ImportFieldProfileService struct {
	repo *repository.ImportFieldProfileRepository
}

func NewImportFieldProfileService(repo *repository.ImportFieldProfileRepository) *ImportFieldProfileService {
	return &ImportFieldProfileService{repo: repo}
}

func normalizeScope(scope string) string {
	return strings.TrimSpace(scope)
}

func normalizeCode(code string) string {
	return strings.TrimSpace(strings.ToLower(code))
}

func (s *ImportFieldProfileService) List(ctx context.Context, scope string) ([]model.ImportFieldProfile, error) {
	return s.repo.List(ctx, normalizeScope(scope))
}

func (s *ImportFieldProfileService) Get(ctx context.Context, id int) (*model.ImportFieldProfile, error) {
	return s.repo.Get(ctx, id)
}

func (s *ImportFieldProfileService) Create(ctx context.Context, in model.ImportFieldProfileInput) (*model.ImportFieldProfile, error) {
	scope := normalizeScope(in.Scope)
	code := normalizeCode(in.Code)
	name := strings.TrimSpace(in.Name)

	if scope == "" {
		return nil, fmt.Errorf("scope is required")
	}
	if code == "" {
		return nil, fmt.Errorf("code is required")
	}
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}

	p := &model.ImportFieldProfile{
		Scope:       scope,
		Code:        code,
		Name:        name,
		Description: nil,
		IsDefault:   in.IsDefault,
		PivotField:  in.PivotField,
		Permission:  in.Permission,
	}
	if in.Description != nil && strings.TrimSpace(*in.Description) != "" {
		s := strings.TrimSpace(*in.Description)
		p.Description = &s
	}

	if p.IsDefault {
		if err := s.repo.UnsetDefaultByScope(ctx, scope); err != nil {
			return nil, err
		}
	}

	created, err := s.repo.Create(ctx, p)
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (s *ImportFieldProfileService) Update(ctx context.Context, id int, in model.ImportFieldProfileInput) (*model.ImportFieldProfile, error) {
	cur, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	if scope := normalizeScope(in.Scope); scope != "" {
		cur.Scope = scope
	}
	if code := normalizeCode(in.Code); code != "" {
		cur.Code = code
	}
	if name := strings.TrimSpace(in.Name); name != "" {
		cur.Name = name
	}
	if in.PivotField != nil && strings.TrimSpace(*in.PivotField) != "" {
		pf := strings.TrimSpace(*in.PivotField)
		cur.PivotField = &pf
	} else {
		cur.PivotField = nil
	}
	if in.Permission != nil && strings.TrimSpace(*in.Permission) != "" {
		p := strings.TrimSpace(*in.Permission)
		cur.Permission = &p
	} else {
		cur.Permission = nil
	}
	if in.Description != nil {
		if strings.TrimSpace(*in.Description) == "" {
			cur.Description = nil
		} else {
			sdesc := strings.TrimSpace(*in.Description)
			cur.Description = &sdesc
		}
	}
	cur.IsDefault = in.IsDefault

	if cur.IsDefault {
		if err := s.repo.UnsetDefaultByScope(ctx, cur.Scope); err != nil {
			return nil, err
		}
	}

	updated, err := s.repo.Update(ctx, cur)
	if err != nil {
		return nil, err
	}
	return updated, nil
}

func (s *ImportFieldProfileService) Delete(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}
