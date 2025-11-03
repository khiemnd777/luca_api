package utils

import (
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// GetUserID returns user ID from Fiber context as string (supports int, int64, string)
func GetUserID(c *fiber.Ctx) (string, bool) {
	val := c.Locals("userID")
	switch v := val.(type) {
	case string:
		return v, true
	case int:
		return fmt.Sprintf("%d", v), true
	case int64:
		return fmt.Sprintf("%d", v), true
	default:
		return "", false
	}
}

// GetUserIDInt returns user ID as int from Fiber context (fallback 0)
func GetUserIDInt(c *fiber.Ctx) (int, bool) {
	val := c.Locals("userID")
	switch v := val.(type) {
	case int:
		return v, true
	case int64:
		return int(v), true
	case string:
		i, err := strconv.Atoi(v)
		if err == nil {
			return i, true
		}
	}
	return 0, false
}

// GetUserRole returns the user's role from context (assumed string)
func GetUserRole(c *fiber.Ctx) string {
	if role, ok := c.Locals("role").(string); ok {
		return role
	}
	return ""
}

// GetUserEmail returns the user's email from context (assumed string)
func GetUserEmail(c *fiber.Ctx) string {
	if email, ok := c.Locals("email").(string); ok {
		return email
	}
	return ""
}

// GetUserWithPermission returns both user ID and permission from context
func GetUserWithPermission(c *fiber.Ctx) (string, string) {
	userID, _ := GetUserID(c)
	perm := ""
	if p, ok := c.Locals("permission").(string); ok {
		perm = p
	}
	return userID, perm
}
