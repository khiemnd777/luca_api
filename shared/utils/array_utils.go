package utils

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
