package utils

import (
	"errors"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/khiemnd777/andy_api/shared/app/client_error"
)

type JWTTokenPayload struct {
	UserID       int
	DepartmentID int
	Email        string
	Roles        *map[string]struct{}
	Permissions  *map[string]struct{}
	Exp          time.Time
}

func GenerateJWTToken(secret string, payload JWTTokenPayload) (string, error) {
	claims := jwt.MapClaims{
		"user_id": payload.UserID,
		"dept_id": payload.DepartmentID,
		"email":   payload.Email,
		"roles":   payload.Roles,
		"perms":   payload.Permissions,
		"exp":     payload.Exp.Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func GetJWTClaims(c *fiber.Ctx) (jwt.MapClaims, bool, error) {
	secret := GetAuthSecret()
	header := c.Get("Authorization")
	if header == "" || !strings.HasPrefix(header, "Bearer ") {
		return nil, false, client_error.ResponseError(c, fiber.StatusUnauthorized, nil, "Missing or invalid Authorization header")
	}

	tokenStr := strings.TrimPrefix(header, "Bearer ")
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return nil, false, client_error.ResponseError(c, fiber.StatusUnauthorized, err, "Invalid or expired token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || claims["user_id"] == nil {
		return nil, false, client_error.ResponseError(c, fiber.StatusUnauthorized, err, "Invalid token claims")
	}

	return claims, true, nil
}

func GetPermSetFromClaims(c *fiber.Ctx) (map[string]struct{}, bool) {
	if v := c.Locals("permSet"); v != nil {
		switch vv := v.(type) {
		case map[string]struct{}:
			if len(vv) > 0 {
				return vv, true
			}
		case *map[string]struct{}:
			if vv != nil && len(*vv) > 0 {
				return *vv, true
			}
		}
	}

	claims, ok, _ := GetJWTClaims(c)
	if !ok || claims == nil {
		return nil, false
	}
	raw, exists := claims["perms"]
	if !exists || raw == nil {
		return nil, false
	}

	normalize := func(s string) string {
		return strings.ToLower(strings.TrimSpace(s))
	}

	set := make(map[string]struct{})

	switch v := raw.(type) {
	case *map[string]struct{}:
		if v != nil {
			for k := range *v {
				if k = normalize(k); k != "" {
					set[k] = struct{}{}
				}
			}
		}
	case map[string]struct{}:
		for k := range v {
			if k = normalize(k); k != "" {
				set[k] = struct{}{}
			}
		}
	case map[string]bool:
		for k, val := range v {
			if !val {
				continue
			}
			if k = normalize(k); k != "" {
				set[k] = struct{}{}
			}
		}
	case map[string]any: // JWT decode JSON object -> map[string]interface{}
		for k := range v {
			if k = normalize(k); k != "" {
				set[k] = struct{}{}
			}
		}
	case []string:
		for _, k := range v {
			if k = normalize(k); k != "" {
				set[k] = struct{}{}
			}
		}
	case []any:
		for _, it := range v {
			if s, ok := it.(string); ok {
				if s = normalize(s); s != "" {
					set[s] = struct{}{}
				}
			}
		}
	case string:
		s := strings.TrimSpace(v)
		var parts []string
		if strings.Contains(s, ",") {
			for _, p := range strings.Split(s, ",") {
				parts = append(parts, p)
			}
		} else {
			parts = strings.Fields(s)
		}
		for _, p := range parts {
			if p = normalize(p); p != "" {
				set[p] = struct{}{}
			}
		}
	default:
		return nil, false
	}

	if len(set) == 0 {
		return nil, false
	}
	c.Locals("permSet", set)
	return set, true
}
