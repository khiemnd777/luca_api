package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/khiemnd777/andy_api/modules/metadata/model"
	"github.com/khiemnd777/andy_api/modules/metadata/repository"
)

type ImportFieldMappingService struct {
	maps    *repository.ImportFieldMappingRepository
	profile *repository.ImportFieldProfileRepository
}

func NewImportFieldMappingService(
	maps *repository.ImportFieldMappingRepository,
	profile *repository.ImportFieldProfileRepository,
) *ImportFieldMappingService {
	return &ImportFieldMappingService{
		maps:    maps,
		profile: profile,
	}
}

func normalizeKind(kind string) (string, error) {
	k := strings.ToLower(strings.TrimSpace(kind))
	switch k {
	case "core", "metadata", "external":
		return k, nil
	default:
		return "", fmt.Errorf("invalid internal_kind: %s", kind)
	}
}

func (s *ImportFieldMappingService) ListByProfileID(ctx context.Context, profileID int) ([]model.ImportFieldMapping, error) {
	if profileID <= 0 {
		return nil, fmt.Errorf("profile_id is required")
	}
	return s.maps.ListByProfileID(ctx, profileID)
}

func (s *ImportFieldMappingService) Get(ctx context.Context, id int) (*model.ImportFieldMapping, error) {
	return s.maps.Get(ctx, id)
}

func (s *ImportFieldMappingService) Create(ctx context.Context, in model.ImportFieldMappingInput) (*model.ImportFieldMapping, error) {
	if in.ProfileID <= 0 {
		return nil, fmt.Errorf("profile_id is required")
	}

	// ensure profile exists
	if _, err := s.profile.Get(ctx, in.ProfileID); err != nil {
		return nil, fmt.Errorf("profile not found")
	}

	kind, err := normalizeKind(in.InternalKind)
	if err != nil {
		return nil, err
	}

	path := strings.TrimSpace(in.InternalPath)
	label := strings.TrimSpace(in.InternalLabel)
	if path == "" {
		return nil, fmt.Errorf("internal_path is required")
	}
	if label == "" {
		return nil, fmt.Errorf("internal_label is required")
	}

	dataType := strings.TrimSpace(in.DataType)

	m := &model.ImportFieldMapping{
		ProfileID:     in.ProfileID,
		InternalKind:  kind,
		InternalPath:  path,
		InternalLabel: label,
		DataType:      dataType,
		Required:      in.Required,
		Unique:        in.Unique,
	}

	if in.MetadataCollectionSlug != nil && strings.TrimSpace(*in.MetadataCollectionSlug) != "" {
		slug := strings.TrimSpace(*in.MetadataCollectionSlug)
		m.MetadataCollectionSlug = &slug
	}
	if in.MetadataFieldName != nil && strings.TrimSpace(*in.MetadataFieldName) != "" {
		fn := strings.TrimSpace(*in.MetadataFieldName)
		m.MetadataFieldName = &fn
	}
	if in.ExcelHeader != nil && strings.TrimSpace(*in.ExcelHeader) != "" {
		h := strings.TrimSpace(*in.ExcelHeader)
		m.ExcelHeader = &h
	}
	if in.ExcelColumn != nil && *in.ExcelColumn > 0 {
		col := *in.ExcelColumn
		m.ExcelColumn = &col
	}
	if in.TransformHint != nil && strings.TrimSpace(*in.TransformHint) != "" {
		th := strings.TrimSpace(*in.TransformHint)
		m.TransformHint = &th
	}

	created, err := s.maps.Create(ctx, m)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func (s *ImportFieldMappingService) Update(ctx context.Context, id int, in model.ImportFieldMappingInput) (*model.ImportFieldMapping, error) {
	cur, err := s.maps.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	if in.ProfileID > 0 && in.ProfileID != cur.ProfileID {
		if _, err := s.profile.Get(ctx, in.ProfileID); err != nil {
			return nil, fmt.Errorf("profile not found")
		}
		cur.ProfileID = in.ProfileID
	}

	if strings.TrimSpace(in.InternalKind) != "" {
		kind, err := normalizeKind(in.InternalKind)
		if err != nil {
			return nil, err
		}
		cur.InternalKind = kind
	}
	if strings.TrimSpace(in.InternalPath) != "" {
		cur.InternalPath = strings.TrimSpace(in.InternalPath)
	}
	if strings.TrimSpace(in.InternalLabel) != "" {
		cur.InternalLabel = strings.TrimSpace(in.InternalLabel)
	}
	if strings.TrimSpace(in.DataType) != "" {
		cur.DataType = strings.TrimSpace(in.DataType)
	}

	cur.Required = in.Required
	cur.Unique = in.Unique

	if in.MetadataCollectionSlug != nil {
		if strings.TrimSpace(*in.MetadataCollectionSlug) == "" {
			cur.MetadataCollectionSlug = nil
		} else {
			slug := strings.TrimSpace(*in.MetadataCollectionSlug)
			cur.MetadataCollectionSlug = &slug
		}
	}
	if in.MetadataFieldName != nil {
		if strings.TrimSpace(*in.MetadataFieldName) == "" {
			cur.MetadataFieldName = nil
		} else {
			fn := strings.TrimSpace(*in.MetadataFieldName)
			cur.MetadataFieldName = &fn
		}
	}
	if in.ExcelHeader != nil {
		if strings.TrimSpace(*in.ExcelHeader) == "" {
			cur.ExcelHeader = nil
		} else {
			h := strings.TrimSpace(*in.ExcelHeader)
			cur.ExcelHeader = &h
		}
	}
	if in.ExcelColumn != nil {
		if *in.ExcelColumn <= 0 {
			cur.ExcelColumn = nil
		} else {
			col := *in.ExcelColumn
			cur.ExcelColumn = &col
		}
	}
	if in.TransformHint != nil {
		if strings.TrimSpace(*in.TransformHint) == "" {
			cur.TransformHint = nil
		} else {
			th := strings.TrimSpace(*in.TransformHint)
			cur.TransformHint = &th
		}
	}

	updated, err := s.maps.Update(ctx, cur)
	if err != nil {
		return nil, err
	}
	return updated, nil
}

func (s *ImportFieldMappingService) Delete(ctx context.Context, id int) error {
	return s.maps.Delete(ctx, id)
}
