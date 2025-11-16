package customfields

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"time"
)

type ValidateResult struct {
	Clean map[string]any    // dữ liệu sạch để lưu vào JSONB
	Errs  map[string]string // field -> error message
}

func (m *Manager) Validate(ctx context.Context, slug string, incoming map[string]any, isPatch bool) (*ValidateResult, error) {
	schema, err := m.GetSchema(ctx, slug)
	if err != nil {
		return nil, err
	}

	defs := map[string]FieldDef{}
	for _, f := range schema.Fields {
		defs[f.Name] = f
	}

	res := &ValidateResult{Clean: map[string]any{}, Errs: map[string]string{}}

	// 1) Apply defaults (chỉ khi create hoặc PATCH mà field chưa gửi)
	if !isPatch {
		for name, f := range defs {
			if f.DefaultValue != nil {
				res.Clean[name] = f.DefaultValue
			}
		}
	}

	// 2) Duyệt incoming → coerce & validate
	for name, raw := range incoming {
		f, ok := defs[name]
		if !ok {
			// cho phép field lạ? thường là KHÔNG
			res.Errs[name] = ErrUnknownField.Error()
			continue
		}
		val, verr := coerceValue(f, raw)
		if verr != nil {
			res.Errs[name] = verr.Error()
			continue
		}
		// kiểm tra lựa chọn (select/multiselect)
		if e := validateOptions(f, val); e != nil {
			res.Errs[name] = e.Error()
			continue
		}
		res.Clean[name] = val
	}

	// 3) Kiểm tra required (create) hoặc required khi PATCH có key đó
	for name, f := range defs {
		if f.Required {
			if isPatch {
				if _, sent := incoming[name]; sent {
					if _, ok := res.Clean[name]; !ok {
						res.Errs[name] = ErrRequired.Error()
					}
				}
			} else {
				if _, ok := res.Clean[name]; !ok {
					res.Errs[name] = ErrRequired.Error()
				}
			}
		}
	}

	if len(res.Errs) > 0 {
		return res, nil
	}
	return res, nil
}

func coerceValue(f FieldDef, raw any) (any, error) {
	switch f.Type {
	case TypeText, TypeRichText, TypeRelation:
		return fmt.Sprintf("%v", raw), nil
	case TypeNumber:
		switch v := raw.(type) {
		case float64:
			return v, nil
		case int:
			return float64(v), nil
		case int64:
			return float64(v), nil
		case string:
			if v == "" {
				return nil, nil
			}
			n, err := strconv.ParseFloat(v, 64)
			if err != nil {
				return nil, ErrInvalidType
			}
			return n, nil
		default:
			return nil, ErrInvalidType
		}
	case TypeBool:
		switch v := raw.(type) {
		case bool:
			return v, nil
		case string:
			if v == "true" || v == "1" {
				return true, nil
			}
			if v == "false" || v == "0" {
				return false, nil
			}
			return nil, ErrInvalidType
		default:
			return nil, ErrInvalidType
		}
	case TypeDate:
		// hỗ trợ YYYY-MM-DD hoặc RFC3339
		s := fmt.Sprintf("%v", raw)
		if _, err := time.Parse(time.RFC3339, s); err == nil {
			return s, nil
		}
		if ok, _ := regexp.MatchString(`^\d{4}-\d{2}-\d{2}$`, s); ok {
			return s, nil
		}
		return nil, ErrInvalidType
	case TypeSelect:
		return fmt.Sprintf("%v", raw), nil
	case TypeMultiSelect:
		switch v := raw.(type) {
		case []any:
			out := make([]string, 0, len(v))
			for _, it := range v {
				out = append(out, fmt.Sprintf("%v", it))
			}
			return out, nil
		default:
			return nil, ErrInvalidType
		}
	case TypeJSON:
		// chấp nhận object/map/list đã parse
		return raw, nil
	default:
		return nil, ErrInvalidType
	}
}

func validateOptions(f FieldDef, v any) error {
	// Choices cho select/multiselect
	if f.Type == TypeSelect || f.Type == TypeMultiSelect {
		if f.Options == nil {
			return nil
		}
		// extract allowed values
		allow := map[string]struct{}{}
		if arr, ok := f.Options["choices"].([]any); ok {
			for _, it := range arr {
				if m, ok := it.(map[string]any); ok {
					if val, ok := m["value"]; ok {
						allow[fmt.Sprintf("%v", val)] = struct{}{}
					}
				}
			}
		}
		if f.Type == TypeSelect {
			s := fmt.Sprintf("%v", v)
			if len(allow) > 0 {
				if _, ok := allow[s]; !ok {
					return ErrInvalidOption
				}
			}
		} else {
			// multi
			list, _ := v.([]string)
			for _, s := range list {
				if len(allow) > 0 {
					if _, ok := allow[s]; !ok {
						return ErrInvalidOption
					}
				}
			}
		}
	}
	// min/max cho number
	if f.Type == TypeNumber && f.Options != nil {
		if v == nil {
			return nil
		}
		n, _ := v.(float64)
		if min, ok := f.Options["min"].(float64); ok && n < min {
			return fmt.Errorf("min %v", min)
		}
		if max, ok := f.Options["max"].(float64); ok && n > max {
			return fmt.Errorf("max %v", max)
		}
	}
	return nil
}

// Merge patch: xóa khóa khi value=nil (nếu cho phép)
func MergePatch(dst map[string]any, patch map[string]any) map[string]any {
	out := map[string]any{}
	for k, v := range dst {
		out[k] = v
	}
	for k, v := range patch {
		if v == nil {
			delete(out, k)
			continue
		}
		out[k] = v
	}
	return out
}
