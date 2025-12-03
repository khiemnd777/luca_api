package customfields

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/khiemnd777/andy_api/shared/utils"
)

func isNilLike(v any) bool {
	if v == nil {
		return true
	}

	switch t := v.(type) {
	case string:
		return t == ""
	case []any:
		return len(t) == 0
	case map[string]any:
		return len(t) == 0
	}

	// reflect nil (pointer, slice, map…)
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Ptr, reflect.Map, reflect.Slice, reflect.Interface:
		return rv.IsNil()
	}

	return false
}

func EvaluateShowIf(cond *ShowIfCondition, data map[string]any) bool {
	if cond == nil {
		return true
	}

	// --- ALL ---
	if len(cond.All) > 0 {
		for _, c := range cond.All {
			cc := c
			if !EvaluateShowIf(&cc, data) {
				return false
			}
		}
		return true
	}

	// --- ANY ---
	if len(cond.Any) > 0 {
		for _, c := range cond.Any {
			cc := c
			if EvaluateShowIf(&cc, data) {
				return true
			}
		}
		return false
	}

	// --- SINGLE CONDITION ---
	v := LookupNestedField(data, cond.Field)

	// normalize bool
	boolVal := toBool(v)

	switch cond.Op {

	// =============================
	// BOOLEAN SPECIAL CASES
	// =============================

	case "is_true":
		return boolVal

	case "is_false":
		return !boolVal

	case "truthy":
		return isTruthy(v)

	case "falsy":
		return !isTruthy(v)

	case "exists":
		return !isNilLike(v)

	case "not_exists":
		return isNilLike(v)

	// =============================
	// EXISTING OPS
	// =============================

	case "eq", "equals":
		if cond.Value == nil {
			return isNilLike(v)
		}
		return fmt.Sprint(v) == fmt.Sprint(cond.Value)

	case "neq", "not_equals":
		if cond.Value == nil {
			return !isNilLike(v)
		}
		return fmt.Sprint(v) != fmt.Sprint(cond.Value)

	case "in":
		return utils.ValueInList(v, cond.Value)

	case "gt":
		return utils.ToFloat(v) > utils.ToFloat(cond.Value)

	case "lt":
		return utils.ToFloat(v) < utils.ToFloat(cond.Value)
	}

	return false
}

func toBool(v any) bool {
	switch t := v.(type) {
	case bool:
		return t
	case int, int8, int16, int32, int64:
		return utils.ToFloat(t) != 0
	case float32, float64:
		return utils.ToFloat(t) != 0
	case string:
		s := strings.TrimSpace(strings.ToLower(t))
		return s == "true" || s == "1" || s == "yes"
	}

	return false
}

func isTruthy(v any) bool {
	if isNilLike(v) {
		return false
	}
	switch t := v.(type) {
	case bool:
		return t
	case string:
		s := strings.TrimSpace(strings.ToLower(t))
		return s != "" && s != "0" && s != "false" && s != "no"
	}
	return true
}
