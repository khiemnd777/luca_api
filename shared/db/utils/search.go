package dbutils

import (
	"context"
	"strings"

	"entgo.io/ent/dialect/sql"
	"github.com/gofiber/fiber/v2"
	"github.com/iancoleman/strcase"
	"github.com/khiemnd777/andy_api/shared/utils"
)

type SearchResult[T any] struct {
	Items   []*T `json:"items"`
	HasMore bool `json:"has_more"`
	Total   int  `json:"total"`
}

type SearchQuery struct {
	Keyword   string  `json:"keyword"`
	Limit     int     `json:"limit"`
	Page      int     `json:"page"`
	Offset    int     `json:"offset"`
	OrderBy   *string `json:"order_by"`
	Direction string  `json:"direction"`
}

func ParseSearchQuery(c *fiber.Ctx, defLimit int) SearchQuery {
	kw := utils.GetQueryAsString(c, "keyword")
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

	if orderBy != "" {
		orderBy = strcase.ToSnake(orderBy)
	}

	direction := utils.GetQueryAsString(c, "direction")
	if direction == "" {
		direction = "asc"
	}

	return SearchQuery{
		Keyword:   kw,
		Limit:     limit,
		Page:      page,
		Offset:    offset,
		OrderBy:   &orderBy,
		Direction: direction,
	}
}

func isDesc(dir string) bool { return strings.EqualFold(dir, "desc") }

func resolveOrderField(orderBy *string, defaultField string) (field string) {
	field = defaultField
	if orderBy != nil && strings.TrimSpace(*orderBy) != "" {
		field = strings.TrimSpace(*orderBy)
	}
	return field
}

// O ~ func(*sql.Selector) để khớp với <entity>.OrderOption
func buildSQLOptions[O ~func(*sql.Selector)](table, field string, desc bool, pkField string) []O {
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

// ================= Generic Search =================

// Q: *<Entity>Query      E: <Entity> (non-pointer)     P: predicate.<Entity> (~func(*sql.Selector))
// O: <entity>.OrderOption (~func(*sql.Selector))
type EntQueryLike[
	E any,
	P ~func(*sql.Selector),
	O ~func(*sql.Selector),
	Q any,
] interface {
	Clone() Q
	Count(context.Context) (int, error)
	Where(...P) Q
	Limit(int) Q
	Offset(int) Q
	Order(...O) Q
	All(context.Context) ([]*E, error)
}

// KwPred: builder predicate phụ thuộc keyword (HasXWith(...), Or lồng…)
type KwPred[P ~func(*sql.Selector)] func(norm string) P

// Search: gom full logic search + order + mapper
// - q:            Query đã Where(...) base (vd: DeletedAtIsNil())
// - likeColumns:  danh sách cột LIKE (đã *_norm nếu cần)
// - sq:           SearchQuery (limit, offset, order_by, direction…)
// - table:        tên bảng Ent (vd: product.Table) — để qualify order column nếu cần
// - pkField:      field PK cho tie-breaker (vd: product.FieldID)
// - defaultField: field mặc định khi order_by rỗng (vd: product.FieldCreatedAt)
// - orFn:         hàm Or(...) của entity (vd: product.Or)
// - mapper:       optional mapper: []*E -> []*R (nếu nil, cast thẳng sang []*R)
// - extras:       các KwPred theo keyword (HasUserWith(... LikeNorm[predicate.User](...)))
// if R=E -> Search[R,R](..., func mapper = nil)
// if R≠E → Search(...)
func Search[
	E any,
	R any,
	P ~func(*sql.Selector),
	O ~func(*sql.Selector),
	Q EntQueryLike[E, P, O, Q],
](
	ctx context.Context,
	q Q,
	likeColumns []string,
	sq SearchQuery,
	table string,
	pkField string,
	defaultField string,
	orFn func(...P) P,
	mapper func(src []*E) []*R,
	extras ...KwPred[P],
) (SearchResult[R], error) {
	norm := utils.NormalizeSearchKeyword(sq.Keyword)
	if norm == "" {
		return SearchResult[R]{}, nil
	}

	total, err := q.Clone().Count(ctx)
	if err != nil {
		return SearchResult[R]{}, err
	}

	// WHERE ... OR (likes..., extras...)
	preds := LikeNormWithMultiColumns[P](norm, likeColumns...)
	for _, f := range extras {
		preds = append(preds, f(norm))
	}
	q = q.Where(orFn(preds...))

	// ORDER BY
	field := resolveOrderField(sq.OrderBy, defaultField)
	desc := isDesc(sq.Direction)
	orderOpts := buildSQLOptions[O](table, field, desc, pkField)
	q = q.Order(orderOpts...)

	// PAGING (limit + 1 cho hasMore)
	q = q.Limit(sq.Limit + 1).Offset(sq.Offset)

	// FETCH
	srcItems, err := q.All(ctx)
	if err != nil {
		return SearchResult[R]{}, err
	}

	// hasMore
	srcItems, hasMore := TrimHasMore(srcItems, sq.Limit)

	// MAP
	if mapper != nil {
		dstItems := mapper(srcItems)
		return SearchResult[R]{
			Items:   dstItems,
			HasMore: hasMore,
			Total:   total,
		}, nil
	}

	// cast if R == E
	anyItems := make([]*R, len(srcItems))
	for i, it := range srcItems {
		anyItems[i] = any(*it).(*R)
	}
	return SearchResult[R]{
		Items:   anyItems,
		HasMore: hasMore,
		Total:   total,
	}, nil
}
