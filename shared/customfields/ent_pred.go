package customfields

import (
	"entgo.io/ent/dialect/sql"
)

// Ví dụ: q.Where(customfields.JSONEq("color", "red"))
func JSONEq(key string, val any) func(*sql.Selector) {
	return func(s *sql.Selector) {
		s.Where(sql.ExprP(" (custom_fields->>?) = ? ", key, val))
	}
}

func JSONILike(key string, pattern string) func(*sql.Selector) {
	return func(s *sql.Selector) {
		s.Where(sql.ExprP(" (custom_fields->>?) ILIKE ? ", key, pattern))
	}
}

func JSONNumOp(key, op string, val any) func(*sql.Selector) {
	return func(s *sql.Selector) {
		s.Where(sql.ExprP(" (custom_fields->>?)::numeric "+op+" ? ", key, val))
	}
}
