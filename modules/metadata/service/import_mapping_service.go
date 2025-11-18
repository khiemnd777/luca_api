package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/khiemnd777/andy_api/modules/metadata/model"
	"github.com/khiemnd777/andy_api/modules/metadata/repository"
	"github.com/khiemnd777/andy_api/shared/middleware/rbac"
	"github.com/khiemnd777/andy_api/shared/utils"
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

func (s *ImportFieldMappingService) ResolveProfileAndMappings(
	c *fiber.Ctx,
	ctx context.Context,
	scope, code string,
) (*model.ImportFieldProfile, []model.ImportFieldMapping, error) {
	var prof *model.ImportFieldProfile
	var err error

	if strings.TrimSpace(code) != "" {
		prof, err = s.profile.GetByScopeAndCode(ctx, scope, code)
	} else {
		prof, err = s.profile.GetDefaultByScope(ctx, scope)
	}
	if err != nil {
		return nil, nil, err
	}

	perms, _ := utils.GetPermSetFromClaims(c)
	profPerms := utils.NormalizeSplit(prof.Permission, ",")
	if rbac.HasAnyPerm(perms, profPerms...) {
		return nil, nil, fmt.Errorf("forbidden: missing permission")
	}

	mappings, err := s.maps.ListByProfileID(ctx, prof.ID)
	if err != nil {
		return nil, nil, err
	}
	if len(mappings) == 0 {
		return nil, nil, fmt.Errorf("no mappings for profile_id=%d", prof.ID)
	}
	return prof, mappings, nil
}

func (s *ImportFieldMappingService) parseValue(dataType, raw string) (any, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}

	switch dataType {
	case "", "text":
		return raw, nil
	case "number":
		if strings.Contains(raw, ".") {
			return strconv.ParseFloat(raw, 64)
		}
		return strconv.ParseInt(raw, 10, 64)
	case "boolean", "bool":
		l := strings.ToLower(raw)
		return l == "1" || l == "true" || l == "yes", nil
	case "date", "datetime":
		return raw, nil
	default:
		return raw, nil
	}
}

func (s *ImportFieldMappingService) MapExcelRow(
	ctx context.Context,
	profileID int,
	valuesByCol map[int]string,
) (*model.MappedRow, error) {
	mappings, err := s.ListByProfileID(ctx, profileID)
	if err != nil {
		return nil, err
	}
	if len(mappings) == 0 {
		return nil, fmt.Errorf("no mappings found for profile_id=%d", profileID)
	}

	row := &model.MappedRow{
		ProfileID:      profileID,
		CoreFields:     map[string]any{},
		MetadataFields: map[string]any{},
		ExternalFields: map[string]any{},
		Cells:          []model.MappedCell{},
	}

	for _, m := range mappings {
		if m.ExcelColumn == nil || *m.ExcelColumn <= 0 {
			continue
		}

		colIdx := *m.ExcelColumn
		raw := strings.TrimSpace(valuesByCol[colIdx])

		cell := model.MappedCell{
			InternalKind:  m.InternalKind,
			InternalPath:  m.InternalPath,
			InternalLabel: m.InternalLabel,
			DataType:      m.DataType,
			RawValue:      raw,
		}

		// parse value theo data_type
		val, _ := s.parseValue(m.DataType, raw)
		cell.Value = val

		// 3) đổ vào các map tương ứng
		switch m.InternalKind {
		case "core":
			row.CoreFields[m.InternalPath] = val
		case "metadata":
			row.MetadataFields[m.InternalPath] = val
		case "external":
			row.ExternalFields[m.InternalPath] = val
		default:
			// unknown kind -> bỏ qua, hoặc log nếu cần
		}

		row.Cells = append(row.Cells, cell)
	}

	return row, nil
}

type ExcelRow struct {
	Columns map[int]any // key: column index 1-based; val: value
}

func (s *ImportFieldMappingService) BuildHeaderRow(
	ctx context.Context,
	profileID int,
) (*ExcelRow, error) {
	mappings, err := s.ListByProfileID(ctx, profileID)
	if err != nil {
		return nil, err
	}
	row := &ExcelRow{Columns: map[int]any{}}

	for _, m := range mappings {
		if m.ExcelColumn == nil || *m.ExcelColumn <= 0 {
			continue
		}
		col := *m.ExcelColumn
		header := m.ExcelHeader
		if header == nil || strings.TrimSpace(*header) == "" {
			// fallback: internal_label
			h := m.InternalLabel
			header = &h
		}
		row.Columns[col] = *header
	}
	return row, nil
}

// entityData: gom core + metadata + external thành map (tuỳ bạn)
type EntityData struct {
	Core     map[string]any
	Metadata map[string]any
	External map[string]any
}

func (s *ImportFieldMappingService) BuildDataRow(
	ctx context.Context,
	profileID int,
	entity *EntityData,
) (*ExcelRow, error) {
	mappings, err := s.ListByProfileID(ctx, profileID)
	if err != nil {
		return nil, err
	}
	row := &ExcelRow{Columns: map[int]any{}}

	for _, m := range mappings {
		if m.ExcelColumn == nil || *m.ExcelColumn <= 0 {
			continue
		}
		col := *m.ExcelColumn

		var v any
		switch m.InternalKind {
		case "core":
			v = entity.Core[m.InternalPath]
		case "metadata":
			v = entity.Metadata[m.InternalPath]
		case "external":
			v = entity.External[m.InternalPath]
		}

		row.Columns[col] = v
	}
	return row, nil
}
