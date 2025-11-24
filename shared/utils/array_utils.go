package utils

import (
	"fmt"
	"reflect"
)

func Dedup[T comparable](input []T, capacity int) []T {
	if capacity <= 0 {
		capacity = len(input)
	}
	result := make([]T, 0, capacity)
	result = append(result, input...)
	if len(result) <= 1 {
		return result
	}

	seen := make(map[T]struct{}, len(result))
	uniq := result[:0]
	for _, v := range result {
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		uniq = append(uniq, v)
	}
	return uniq
}

func DedupInt(input []int, capacity int) []int {
	return Dedup(input, capacity)
}

// e.g.
// ValueInList("A", []string{"A","B"})             // true
// ValueInList(3, []int{1,2,3})                    // true
// ValueInList("3", []int{1,2,3})                  // true
// ValueInList("active", []any{"active","pause"})  // true
// ValueInList(10, []float64{10,20})               // true
// ValueInList(10, nil)                            // false
// ValueInList(5, "abc")                           // false
func ValueInList(v any, list any) bool {
	if list == nil {
		return false
	}

	val := reflect.ValueOf(list)

	kind := val.Kind()
	if kind != reflect.Slice && kind != reflect.Array {
		return false
	}

	for i := 0; i < val.Len(); i++ {
		item := val.Index(i).Interface()

		if reflect.DeepEqual(v, item) {
			return true
		}

		if fmt.Sprint(v) == fmt.Sprint(item) {
			return true
		}
	}

	return false
}
