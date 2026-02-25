package utils

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
)

const (
	maxInt = int(^uint(0) >> 1)
	minInt = -maxInt - 1
)

func GetParamAsNillableInt(c *fiber.Ctx, paramName string) (*int, error) {
	var result *int
	if raw := c.Params(paramName); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			return nil, err
		}
		result = &parsed
	}
	return result, nil
}

func GetParamAsInt(c *fiber.Ctx, paramName string) (int, error) {
	return strconv.Atoi(c.Params(paramName))
}

func GetParamAsString(c *fiber.Ctx, paramName string) string {
	return c.Params(paramName)
}

func GetQueryAsNillableInt(c *fiber.Ctx, queryName string, defaultValue ...string) (*int, error) {
	if raw := c.Query(queryName, defaultValue...); raw != "" {
		parsed, err := strconv.Atoi(raw)

		if err != nil {
			return nil, err
		}
		return &parsed, nil
	}
	return nil, nil
}

func GetQueryAsInt(c *fiber.Ctx, queryName string, defaultValue ...int) int {
	raw := c.Query(queryName)
	if raw == "" {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return 0
	}

	parsed, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || parsed > int64(maxInt) || parsed < int64(minInt) {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return 0
	}

	return int(parsed)
}

func GetQueryAsInt64(c *fiber.Ctx, queryName string, defaultValue ...int64) int64 {
	raw := c.Query(queryName)
	if raw == "" {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return 0
	}

	parsed, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return 0
	}

	return parsed
}

func GetQueryAsString(c *fiber.Ctx, queryName string, defaultValue ...string) string {
	return c.Query(queryName, defaultValue...)
}

func GetQueryAsFloat64Pointer(c *fiber.Ctx, name string) (*float64, error) {
	val := c.Query(name)
	if val == "" {
		return nil, nil
	}
	f, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return nil, err
	}
	return &f, nil
}
