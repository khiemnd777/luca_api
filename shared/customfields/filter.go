package customfields

import (
	"fmt"
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
