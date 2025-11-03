package utils

import (
	"strings"

	"entgo.io/ent/dialect/sql"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/predicate"
)

func buildWildPattern(normalizedKeyword string) string {
	parts := strings.Fields(normalizedKeyword)  // "mang   cut" -> ["mang", "cut"]
	return "%" + strings.Join(parts, "%") + "%" // "mang cut" -> "%mang%cut%"
}

func LikeNorm(column, normalizedKeyword string) func(*sql.Selector) {
	pattern := buildWildPattern(normalizedKeyword)
	return func(s *sql.Selector) {
		s.Where(sql.Like(s.C(column), pattern))
	}
}

func LikeNormUser(column, normalizedKeyword string) predicate.User {
	return func(s *sql.Selector) { LikeNorm(column, normalizedKeyword)(s) }
}
