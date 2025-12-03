package utils

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

//
// -----------------------------------------------------
// Helpers
// -----------------------------------------------------
//

// isEmpty kiểm tra các giá trị "rỗng" để return nil trong SafeParseXPtr
func isEmpty(v any) bool {
	if v == nil {
		return true
	}

	switch x := v.(type) {
	case string:
		return strings.TrimSpace(x) == ""
	case []byte:
		return strings.TrimSpace(string(x)) == ""
	}

	// fallback
	s := strings.TrimSpace(fmt.Sprintf("%v", v))
	return s == "" || s == "<nil>"
}

//
// -----------------------------------------------------
// Base SafeParse (giữ nguyên behavior cũ)
// -----------------------------------------------------
//

func SafeParseString(v any) string {
	if v == nil {
		return ""
	}

	switch x := v.(type) {
	case string:
		return x
	case fmt.Stringer:
		return x.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}

func SafeParseBool(v any) bool {
	switch x := v.(type) {
	case nil:
		return false
	case bool:
		return x
	case int, int64, int32:
		return SafeParseInt(v) != 0
	case float32, float64:
		return SafeParseFloat(v) != 0
	case string:
		s := strings.ToLower(strings.TrimSpace(x))
		return s == "true" || s == "1" || s == "yes" || s == "on"
	default:
		return false
	}
}

func SafeParseFloat(v any) float64 {
	switch x := v.(type) {
	case nil:
		return 0
	case float32:
		return float64(x)
	case float64:
		return x
	case int:
		return float64(x)
	case int8, int16, int32, int64:
		return float64(reflect.ValueOf(x).Int())
	case uint, uint8, uint16, uint32, uint64:
		return float64(reflect.ValueOf(x).Uint())
	case json.Number:
		f, _ := x.Float64()
		return f
	case string:
		f, _ := strconv.ParseFloat(strings.TrimSpace(x), 64)
		return f
	default:
		return 0
	}
}

func SafeParseInt(v any) int {
	switch x := v.(type) {
	case nil:
		return 0
	case int:
		return x
	case int8, int16, int32, int64:
		return int(reflect.ValueOf(x).Int())
	case uint, uint8, uint16, uint32, uint64:
		return int(reflect.ValueOf(x).Uint())
	case float32:
		return int(x)
	case float64:
		return int(x)
	case json.Number:
		i, _ := x.Int64()
		return int(i)
	case string:
		i, _ := strconv.Atoi(strings.TrimSpace(x))
		return i
	default:
		return 0
	}
}

func SafeParseInt64(v any) int64 {
	if v == nil {
		return 0
	}

	switch x := v.(type) {

	case int:
		return int64(x)
	case int8:
		return int64(x)
	case int16:
		return int64(x)
	case int32:
		return int64(x)
	case int64:
		return x

	case uint:
		return int64(x)
	case uint8:
		return int64(x)
	case uint16:
		return int64(x)
	case uint32:
		return int64(x)
	case uint64:
		return int64(x)

	case float32:
		return int64(x)
	case float64:
		return int64(x)

	case json.Number:
		if i, err := x.Int64(); err == nil {
			return i
		}
		if f, err := x.Float64(); err == nil {
			return int64(f)
		}
		return 0

	case string:
		s := strings.TrimSpace(x)
		if s == "" {
			return 0
		}
		if i, err := strconv.ParseInt(s, 10, 64); err == nil {
			return i
		}
		if f, err := strconv.ParseFloat(s, 64); err == nil {
			return int64(f)
		}
		return 0
	}

	return 0
}

//
// -----------------------------------------------------
// SafeParseDateTime (base)
// -----------------------------------------------------
//

func SafeParseDateTime(v any) time.Time {
	var zero time.Time

	if v == nil {
		return zero
	}

	// time.Time
	if t, ok := v.(time.Time); ok {
		return t
	}
	// *time.Time
	if tp, ok := v.(*time.Time); ok {
		if tp == nil {
			return zero
		}
		return *tp
	}

	var s string

	switch x := v.(type) {
	case string:
		s = strings.TrimSpace(x)
	case []byte:
		s = strings.TrimSpace(string(x))
	case json.Number:
		raw := strings.TrimSpace(x.String())
		if raw == "" {
			return zero
		}
		if ts, err := strconv.ParseInt(raw, 10, 64); err == nil {
			return parseTimestampInt(ts)
		}
		s = raw
	default:
		s = strings.TrimSpace(fmt.Sprintf("%v", v))
	}

	if s == "" {
		return zero
	}

	// numeric timestamp
	if i, err := strconv.ParseInt(s, 10, 64); err == nil {
		return parseTimestampInt(i)
	}

	layouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"2006-01-02",
		"02/01/2006 15:04:05",
		"02/01/2006 15:04",
		"02/01/2006",
		"2006/01/02 15:04:05",
		"2006/01/02 15:04",
		"2006/01/02",
	}

	for _, layout := range layouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t
		}
	}

	return zero
}

func parseTimestampInt(v int64) time.Time {
	if v > 1e12 { // miliseconds
		return time.UnixMilli(v)
	}
	if v > 1e9 { // seconds
		return time.Unix(v, 0)
	}
	if v > 1e6 { // micro
		return time.UnixMicro(v)
	}
	return time.Unix(0, v) // nano
}

//
// -----------------------------------------------------
// Pointer Versions — return nil nếu input rỗng
// -----------------------------------------------------
//

func SafeParseStringPtr(v any) *string {
	if isEmpty(v) {
		return nil
	}
	s := SafeParseString(v)
	return &s
}

func SafeParseBoolPtr(v any) *bool {
	if isEmpty(v) {
		return nil
	}
	b := SafeParseBool(v)
	return &b
}

func SafeParseIntPtr(v any) *int {
	if isEmpty(v) {
		return nil
	}
	i := SafeParseInt(v)
	return &i
}

func SafeParseInt64Ptr(v any) *int64 {
	if isEmpty(v) {
		return nil
	}
	i := SafeParseInt64(v)
	return &i
}

func SafeParseFloatPtr(v any) *float64 {
	if isEmpty(v) {
		return nil
	}
	f := SafeParseFloat(v)
	return &f
}

func SafeParseDateTimePtr(v any) *time.Time {
	if isEmpty(v) {
		return nil
	}
	t := SafeParseDateTime(v)
	if t.IsZero() {
		return nil
	}
	return &t
}

//
// -----------------------------------------------------
// SafeGet helpers cho map[string]any
// -----------------------------------------------------
//

// Trả về giá trị any từ map, hoặc nil nếu không có key
func SafeGet(m map[string]any, key string) any {
	if m == nil {
		return nil
	}
	v, ok := m[key]
	if !ok {
		return nil
	}
	return v
}

func SafeGetString(m map[string]any, key string) string {
	return SafeParseString(SafeGet(m, key))
}

func SafeGetStringPtr(m map[string]any, key string) *string {
	return SafeParseStringPtr(SafeGet(m, key))
}

func SafeGetBool(m map[string]any, key string) bool {
	return SafeParseBool(SafeGet(m, key))
}

func SafeGetBoolPtr(m map[string]any, key string) *bool {
	return SafeParseBoolPtr(SafeGet(m, key))
}

func SafeGetInt(m map[string]any, key string) int {
	return SafeParseInt(SafeGet(m, key))
}

func SafeGetIntPtr(m map[string]any, key string) *int {
	return SafeParseIntPtr(SafeGet(m, key))
}

func SafeGetInt64(m map[string]any, key string) int64 {
	return SafeParseInt64(SafeGet(m, key))
}

func SafeGetInt64Ptr(m map[string]any, key string) *int64 {
	return SafeParseInt64Ptr(SafeGet(m, key))
}

func SafeGetFloat(m map[string]any, key string) float64 {
	return SafeParseFloat(SafeGet(m, key))
}

func SafeGetFloatPtr(m map[string]any, key string) *float64 {
	return SafeParseFloatPtr(SafeGet(m, key))
}

func SafeGetDateTime(m map[string]any, key string) time.Time {
	return SafeParseDateTime(SafeGet(m, key))
}

func SafeGetDateTimePtr(m map[string]any, key string) *time.Time {
	return SafeParseDateTimePtr(SafeGet(m, key))
}
