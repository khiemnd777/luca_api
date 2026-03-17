package logger

import "context"

type contextKey string

const fieldsContextKey contextKey = "logger_fields"

func ContextWithField(ctx context.Context, key string, value any) context.Context {
	return ContextWithFields(ctx, map[string]any{key: value})
}

func ContextWithFields(ctx context.Context, fields map[string]any) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	current := fieldsFromContext(ctx)
	for k, v := range fields {
		current[k] = v
	}

	return context.WithValue(ctx, fieldsContextKey, current)
}

func fieldsFromContext(ctx context.Context) map[string]any {
	if ctx == nil {
		return map[string]any{}
	}

	raw := ctx.Value(fieldsContextKey)
	if raw == nil {
		return map[string]any{}
	}

	fields, ok := raw.(map[string]any)
	if !ok {
		return map[string]any{}
	}

	return cloneMap(fields)
}
