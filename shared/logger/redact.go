package logger

import "strings"

func redactFields(fields map[string]any, redactKeys []string) map[string]any {
	if len(fields) == 0 {
		return fields
	}

	set := make(map[string]struct{}, len(redactKeys))
	for _, key := range redactKeys {
		normalized := normalizeKey(key)
		if normalized != "" {
			set[normalized] = struct{}{}
		}
	}

	return redactMap(fields, set)
}

func redactMap(fields map[string]any, redactKeys map[string]struct{}) map[string]any {
	out := make(map[string]any, len(fields))
	for k, v := range fields {
		if _, ok := redactKeys[normalizeKey(k)]; ok {
			out[k] = "***redacted***"
			continue
		}
		out[k] = redactValue(v, redactKeys)
	}
	return out
}

func redactValue(value any, redactKeys map[string]struct{}) any {
	switch vv := value.(type) {
	case map[string]any:
		return redactMap(vv, redactKeys)
	case map[string]string:
		out := make(map[string]any, len(vv))
		for k, v := range vv {
			if _, ok := redactKeys[normalizeKey(k)]; ok {
				out[k] = "***redacted***"
				continue
			}
			out[k] = v
		}
		return out
	case []any:
		out := make([]any, 0, len(vv))
		for _, item := range vv {
			out = append(out, redactValue(item, redactKeys))
		}
		return out
	default:
		return value
	}
}

func normalizeKey(key string) string {
	key = strings.ToLower(strings.TrimSpace(key))
	key = strings.ReplaceAll(key, "-", "_")
	return key
}
