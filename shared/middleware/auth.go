package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/khiemnd777/andy_api/shared/app/client_error"
	"github.com/khiemnd777/andy_api/shared/utils"
)

func RequireAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		header := c.Get("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			return client_error.ResponseError(c, fiber.StatusUnauthorized, nil, "Missing or invalid Authorization header")
		}

		tokenStr := strings.TrimPrefix(header, "Bearer ")

		claims, ok, err := utils.GetJWTClaims(c)

		if !ok || err != nil {
			return client_error.ResponseError(c, fiber.StatusUnauthorized, err, "Invalid token claims")
		}

		c.Locals("userID", int(claims["user_id"].(float64)))
		c.Locals("deptID", int(claims["dept_id"].(float64)))

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
