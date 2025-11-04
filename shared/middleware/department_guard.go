package middleware

import (
	"context"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/khiemnd777/andy_api/shared/utils"
)

type DepartmentChecker interface {
	IsMember(ctx context.Context, userID, deptID int) (bool, error)
}

func RequireDepartmentMember(checker DepartmentChecker, deptIDFromPathParam string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, ok := utils.GetUserIDInt(c)
		if !ok || userID <= 0 {
			return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
		}
		deptStr := c.Params(deptIDFromPathParam)
		deptID, err := strconv.Atoi(deptStr)
		if err != nil || deptID <= 0 {
			return fiber.NewError(fiber.StatusBadRequest, "invalid department id")
		}

		ok, err = checker.IsMember(c.UserContext(), userID, deptID)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
		if !ok {
			return fiber.NewError(fiber.StatusForbidden, "forbidden: not a member of department")
		}
		return c.Next()
	}
}
