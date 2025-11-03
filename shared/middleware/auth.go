package middleware

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/khiemnd777/andy_api/shared/app/client_error"
	"github.com/khiemnd777/andy_api/shared/utils"
)

func RequireAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		secret := utils.GetAuthSecret()
		header := c.Get("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			return client_error.ResponseError(c, fiber.StatusUnauthorized, nil, "Missing or invalid Authorization header")
		}

		tokenStr := strings.TrimPrefix(header, "Bearer ")
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return []byte(secret), nil
		})
		if err != nil || !token.Valid {
			return client_error.ResponseError(c, fiber.StatusUnauthorized, err, "Invalid or expired token")
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || claims["user_id"] == nil {
			return client_error.ResponseError(c, fiber.StatusUnauthorized, err, "Invalid token claims")
		}

		c.Locals("userID", int(claims["user_id"].(float64)))

		// Inject token into context for downstream access
		ctxWithToken := utils.SetAccessTokenIntoContext(c.UserContext(), tokenStr)
		c.SetUserContext(ctxWithToken)

		return c.Next()
	}
}

func RequireInternal() fiber.Handler {
	// Only use for internal audit, trace, impersonate, và trust-based routing.
	return func(c *fiber.Ctx) error {
		token := c.Get("X-Internal-Token")
		baseIntrTkn := utils.GetInternalToken()
		if token != baseIntrTkn {
			return c.Status(401).SendString("Unauthorized internal call")
		}
		return c.Next()
	}
}
