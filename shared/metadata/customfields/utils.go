package customfields

import (
	"fmt"

	"github.com/khiemnd777/andy_api/shared/utils"
)

func EvaluateShowIf(cond *ShowIfCondition, data map[string]any) bool {
	if cond == nil {
		return true
	}

	if len(cond.All) > 0 {
		for _, c := range cond.All {
			cc := c
			if !EvaluateShowIf(&cc, data) {
				return false
			}
		}
		return true
	}

	if len(cond.Any) > 0 {
		for _, c := range cond.Any {
			cc := c
			if EvaluateShowIf(&cc, data) {
				return true
			}
		}
		return false
	}

	v := LookupNestedField(data, cond.Field)

	switch cond.Op {
	case "eq", "equals":
		return fmt.Sprint(v) == fmt.Sprint(cond.Value)
	case "neq", "not_equals":
		return fmt.Sprint(v) != fmt.Sprint(cond.Value)
	case "in":
		return utils.ValueInList(v, cond.Value)
	case "gt":
		return utils.ToFloat(v) > utils.ToFloat(cond.Value)
	case "lt":
		return utils.ToFloat(v) < utils.ToFloat(cond.Value)
	default:
		return false
	}
}
