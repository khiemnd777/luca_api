package utils

import (
	"context"
)

type ctxKey string

const CtxAccessTokenKey ctxKey = "accessToken"

// SetAccessTokenIntoContext returns a new context with the access token
func SetAccessTokenIntoContext(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, CtxAccessTokenKey, token)
}

// GetAccessTokenFromContext extracts access token from context.Context
func GetAccessTokenFromContext(ctx context.Context) string {
	if v := ctx.Value(CtxAccessTokenKey); v != nil {
		if token, ok := v.(string); ok {
			return token
		}
	}
	return ""
}
