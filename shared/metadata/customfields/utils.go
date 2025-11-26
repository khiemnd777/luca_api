package customfields

import (
	"fmt"
	"reflect"

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

	switch cond.Op {
	case "eq", "equals":
		// special case eq nil
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
