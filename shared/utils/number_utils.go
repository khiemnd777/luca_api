package utils

import (
	"fmt"
	"strconv"
)

func ParseFloat(s string) float64 {
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return v
}

func ToFloat(v any) float64 {
	if v == nil {
		return 0
	}

	switch x := v.(type) {
	case float64:
		return x
	case float32:
		return float64(x)
	case int:
		return float64(x)
	case int64:
		return float64(x)
	case int32:
		return float64(x)
	case uint:
		return float64(x)
	case uint64:
		return float64(x)
	case uint32:
		return float64(x)
	case string:
		if x == "" {
			return 0
		}
		if f, err := strconv.ParseFloat(x, 64); err == nil {
			return f
		}
		return 0
	default:
		if s, ok := v.(fmt.Stringer); ok {
			f := ParseFloat(s.String())
			return f
		}
		f := ParseFloat(fmt.Sprint(v))
		return f
	}
}
