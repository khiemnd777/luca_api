package table

import (
	"context"
	"fmt"
	"strings"

	"entgo.io/ent/dialect/sql"
	"github.com/gofiber/fiber/v2"
	"github.com/khiemnd777/andy_api/shared/utils"
)

type TableQuery struct {
	Limit     int     `json:"limit"`
	Page      int     `json:"page"`
	Offset    int     `json:"offset"`
	OrderBy   *string `json:"order_by"`
	Direction string  `json:"direction"`
}

type TableListResult[T any] struct {
	Items []*T `json:"items"`
	Total int  `json:"total"`
}

func ParseTableQuery(c *fiber.Ctx, defLimit int) TableQuery {
	limit := utils.GetQueryAsInt(c, "limit", defLimit)
	page := utils.GetQueryAsInt(c, "page", 1)

	if limit <= 0 {
		limit = defLimit
	}
	if page <= 0 {
		page = 1
	}

	offset := (page - 1) * limit

	orderBy := utils.GetQueryAsString(c, "order_by")

	direction := utils.GetQueryAsString(c, "direction")
	if direction == "" {
		direction = "asc"
	}

	return TableQuery{
		Limit:     limit,
		Page:      page,
		Offset:    offset,
		OrderBy:   &orderBy,
		Direction: direction,
	}
}

const (
	DefaultLimit = 20
	MaxLimit     = 200
)

func BuildOrderSQL(q TableQuery) string {
	if q.OrderBy == nil || *q.OrderBy == "" {
		return "ORDER BY id ASC"
	}

	col := utils.ToSnake(*q.OrderBy)

	dir := strings.ToUpper(q.Direction)
	if dir != "ASC" && dir != "DESC" {
		dir = "ASC"
	}

	return fmt.Sprintf("ORDER BY %s %s", col, dir)
}

func BuildLimitSQL(q TableQuery) string {
	limit := q.Limit
	if limit <= 0 {
		limit = 20
	}

	offset := q.Offset
	if offset < 0 {
		offset = 0
	}

	return fmt.Sprintf("LIMIT %d OFFSET %d", limit, offset)
}

// ====== Helpers ======
func normalizePaging(limit, offset int) (int, int) {
	if limit <= 0 {
		limit = DefaultLimit
	}
	if limit > MaxLimit {
		limit = MaxLimit
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}

func resolveOrderField(orderBy *string, defaultField string) (field string, desc bool) {
	field = defaultField
	if orderBy != nil && strings.TrimSpace(*orderBy) != "" {
		field = strings.TrimSpace(*orderBy)
	}
	return field, false
}

func isDesc(dir string) bool { return strings.EqualFold(dir, "desc") }

// Build các OrderOption dựa trên table/field, dùng được cho mọi entity
// O ~ func(*sql.Selector) để khớp với <entity>.OrderOption
func buildSQLOptions[O ~func(*sql.Selector)](table, field string, desc bool, pkField string) []O {
	// custom_fields.<key>
	if strings.HasPrefix(field, "custom_fields.") {
		key := strings.TrimPrefix(field, "custom_fields.")

		// order JSONB: (custom_fields->>'key') [ASC|DESC]
		makeJSON := func(d bool) O {
			return O(func(s *sql.Selector) {
				expr := fmt.Sprintf("(custom_fields->>'%s')", key)
				if d {
					s.OrderBy(expr + " DESC")
				} else {
					s.OrderBy(expr + " ASC")
				}
			})
		}

		// tie-breaker
		makePK := func(d bool) O {
			return O(func(s *sql.Selector) {
				col := s.C(pkField)
				if table != "" {
					col = sql.Table(table).C(pkField)
				}
				if d {
					s.OrderBy(sql.Desc(col))
				} else {
					s.OrderBy(sql.Asc(col))
				}
			})
		}

		opts := []O{makeJSON(desc)}
		if pkField != "" {
			opts = append(opts, makePK(desc))
		}
		return opts
	}

	makeOne := func(f string, d bool) O {
		return O(func(s *sql.Selector) {
			col := s.C(f)
			if table != "" {
				col = sql.Table(table).C(f)
			}
			if d {
				s.OrderBy(sql.Desc(col))
			} else {
				s.OrderBy(sql.Asc(col))
			}
		})
	}

	opts := []O{makeOne(field, desc)}
	if pkField != "" && pkField != field {
		opts = append(opts, makeOne(pkField, desc)) // tie-breaker ổn định
	}
	return opts
}

// ====== Generic cho MỌI bảng Ent ======
// T: entity non-pointer (vd: generated.Role)
// O: order option type của entity đó (vd: role.OrderOption) ~ func(*sql.Selector)
// Q: query type (vd: *generated.RoleQuery)
func TableList[
	T any,
	R any,
	O ~func(*sql.Selector),
	Q interface {
		Clone() Q
		Count(context.Context) (int, error)
		Limit(int) Q
		Offset(int) Q
		Order(...O) Q
		All(context.Context) ([]*T, error)
	},
](
	ctx context.Context,
	q Q,
	opts TableQuery,
	table string, // e.g.: role.Table
	pkField string,
	defaultField string,
	mapItems func(src []*T) []*R, // optional mapper
) (TableListResult[R], error) {

	total, err := q.Clone().Count(ctx)
	if err != nil {
		return TableListResult[R]{}, err
	}

	limit, offset := normalizePaging(opts.Limit, opts.Offset)
	field, _ := resolveOrderField(opts.OrderBy, defaultField)
	desc := isDesc(opts.Direction)
	orderOpts := buildSQLOptions[O](table, field, desc, pkField)

	srcItems, err := q.
		Limit(limit).
		Offset(offset).
		Order(orderOpts...).
		All(ctx)
	if err != nil {
		return TableListResult[R]{}, err
	}

	if mapItems != nil {
		dstItems := mapItems(srcItems)
		return TableListResult[R]{Items: dstItems, Total: total}, nil
	}

	anyItems := make([]*R, len(srcItems))
	for i, item := range srcItems {
		anyItems[i] = any(item).(*R)
	}

	return TableListResult[R]{Items: anyItems, Total: total}, nil
}
