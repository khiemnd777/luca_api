package dbutils

import (
	"fmt"
	"strings"

	"entgo.io/ent/dialect/sql"
	"github.com/khiemnd777/andy_api/shared/utils"
)

func escapeLike(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `%`, `\%`)
	s = strings.ReplaceAll(s, `_`, `\_`)
	return s
}

func buildWildPattern(normalizedKeyword string) string {
	parts := strings.Fields(normalizedKeyword) // "mang   cut" -> ["mang", "cut"]
	for i := range parts {
		parts[i] = escapeLike(parts[i])
	}
	return "%" + strings.Join(parts, "%") + "%" // "mang cut" -> "%mang%cut%"
}

type EntPredicate interface{ ~func(*sql.Selector) }

// Usage:
//
// import: dbutils "github.com/khiemnd777/andy_api/shared/db/utils"
//
// q.Where(utils.LikeNorm[predicate.User]("name", keyword))
func LikeNorm[P EntPredicate](column, normalizedKeyword string) P {
	pattern := buildWildPattern(normalizedKeyword)
	return P(func(s *sql.Selector) {
		s.Where(sql.Like(s.C(column), pattern))
	})
}

func LikeNormWithMultiColumns[P ~func(*sql.Selector)](kw string, columns ...string) []P {
	norm := utils.NormalizeSearchKeyword(kw)
	out := make([]P, 0, len(columns))
	for _, c := range columns {
		out = append(out, LikeNorm[P](c, norm))
	}
	return out
}

func BuildLikeNormSQL(norm string, normColumns []string, args *[]any) string {
	return BuildLikeNormSQLAlias(norm, "r", normColumns, args)
}

func BuildLikeNormSQLAlias(norm string, alias string, normColumns []string, args *[]any) string {
	if norm == "" || len(normColumns) == 0 {
		return ""
	}

	pattern := buildWildPattern(norm)
	idx := len(*args) + 1
	*args = append(*args, pattern)

	parts := make([]string, 0, len(normColumns))
	for _, col := range normColumns {
		target := fmt.Sprintf("%s.%s", alias, col+"_norm")
		if strings.Contains(col, ".") {
			target = fmt.Sprintf("%s%s", col, "_norm")
		}
		parts = append(parts, fmt.Sprintf("%s LIKE $%d", target, idx))
	}

	return "(" + strings.Join(parts, " OR ") + ")"
}

func GetNormField(field string) string {
	return fmt.Sprintf("%s_norm", field)
}

func TrimHasMore[T any](items []*T, limit int) ([]*T, bool) {
	if len(items) > limit {
		return items[:limit], true
	}
	return items, false
}
