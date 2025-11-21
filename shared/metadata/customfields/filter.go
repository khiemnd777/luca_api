package customfields

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// BuildWhereSQL: tạo WHERE snippet + args cho JSONB
// filters: key->value; value có thể: string/bool/float64, hoặc map{"op":"gt","value":10}
func BuildWhereSQL(filters map[string]any, paramStart int) (string, []any) {
	// ví dụ key: "color", "price", "tags"
	i := paramStart
	parts := []string{}
	args := []any{}

	for k, v := range filters {
		op := "="
		cast := ""
		switch t := v.(type) {
		case map[string]any:
			// {op: gt/lt/gte/lte/eq/neq/ilike, value:any}
			if o, ok := t["op"].(string); ok {
				op = strings.ToLower(o)
			}
			v = t["value"]
		}

		// Chọn operator & cast
		switch vv := v.(type) {
		case float64:
			cast = "::numeric"
			if op == "ilike" {
				op = "="
			}
			parts = append(parts, fmt.Sprintf("(custom_fields->>'%s')%s %s $%d", k, cast, opSQL(op), i))
			args = append(args, vv)
		case bool:
			cast = "::boolean"
			parts = append(parts, fmt.Sprintf("(custom_fields->>'%s')%s %s $%d", k, cast, opSQL(op), i))
			args = append(args, vv)
		default:
			if op == "ilike" {
				parts = append(parts, fmt.Sprintf("(custom_fields->>'%s') ILIKE $%d", k, i))
				args = append(args, fmt.Sprintf("%%%v%%", v))
			} else {
				parts = append(parts, fmt.Sprintf("(custom_fields->>'%s') %s $%d", k, opSQL(op), i))
				args = append(args, v)
			}
		}
		i++
	}

	if len(parts) == 0 {
		return "", nil
	}
	return "(" + strings.Join(parts, " AND ") + ")", args
}

func opSQL(op string) string {
	switch op {
	case "gt":
		return ">"
	case "lt":
		return "<"
	case "gte":
		return ">="
	case "lte":
		return "<="
	case "neq":
		return "<>"
	case "eq":
		return "="
	case "ilike":
		return "ILIKE"
	default:
		return "="
	}
}

// LookupNestedField supports:
// - map[string]any
// - map[string]interface{}
// - struct fields
// - array/slice indices (items.0.name)
func LookupNestedField(data any, path string) any {
	if data == nil || path == "" {
		return nil
	}

	parts := strings.Split(path, ".")
	current := data

	for _, key := range parts {

		if current == nil {
			return nil
		}

		val := reflect.ValueOf(current)

		// unwrap pointer
		if val.Kind() == reflect.Ptr {
			if val.IsNil() {
				return nil
			}
			val = val.Elem()
		}

		switch val.Kind() {

		// -----------------------------
		// map[string]any
		// -----------------------------
		case reflect.Map:
			mapKey := reflect.ValueOf(key)
			mv := val.MapIndex(mapKey)
			if !mv.IsValid() {
				return nil
			}
			current = mv.Interface()
			continue

		// -----------------------------
		// slice/array (allow index)
		// -----------------------------
		case reflect.Slice, reflect.Array:
			idx, err := strconv.Atoi(key)
			if err != nil {
				return nil // invalid index
			}
			if idx < 0 || idx >= val.Len() {
				return nil
			}
			current = val.Index(idx).Interface()
			continue

		// -----------------------------
		// struct (public fields)
		// -----------------------------
		case reflect.Struct:
			// Try exact case first: FieldByName("Field")
			f := val.FieldByName(key)
			if f.IsValid() {
				current = f.Interface()
				continue
			}

			// Try case-insensitive match
			typ := val.Type()
			for i := 0; i < typ.NumField(); i++ {
				sf := typ.Field(i)
				if strings.EqualFold(sf.Name, key) {
					current = val.Field(i).Interface()
					goto nextPart
				}
			}

			// Try json tag `json:"field_name"`
			for i := 0; i < typ.NumField(); i++ {
				sf := typ.Field(i)
				if tag := sf.Tag.Get("json"); tag != "" {
					tagName := strings.Split(tag, ",")[0]
					if tagName == key {
						current = val.Field(i).Interface()
						goto nextPart
					}
				}
			}

			return nil

		default:
			return nil
		}

	nextPart:
		continue
	}

	return current
}
