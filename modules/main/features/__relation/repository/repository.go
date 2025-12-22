package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	relation "github.com/khiemnd777/andy_api/modules/main/features/__relation/policy"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	dbutils "github.com/khiemnd777/andy_api/shared/db/utils"
	"github.com/khiemnd777/andy_api/shared/logger"
	"github.com/khiemnd777/andy_api/shared/utils"
	tableutils "github.com/khiemnd777/andy_api/shared/utils/table"
)

type RelationRepository struct{}

func NewRelationRepository() *RelationRepository {
	return &RelationRepository{}
}

func (r *RelationRepository) Get1(
	ctx context.Context,
	tx *generated.Tx,
	cfg relation.Config1,
	id int,
) (any, error) {

	if len(cfg.RefFields) == 0 {
		return nil, fmt.Errorf("relationRepo.Get1: RefFields is empty")
	}

	selectCols, err := buildSelectCols("", cfg.RefFields)
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf(`
        SELECT %s 
        FROM %s
        WHERE %s = $1
        LIMIT 1
    `, selectCols, cfg.RefTable, cfg.RefIDCol)

	rows, err := tx.QueryContext(ctx, sql, id)
	if err != nil {
		return nil, fmt.Errorf("relationRepo.Get1 query: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, nil // không có -> return nil
	}

	scanTargets := buildScanTargets(len(cfg.RefFields))
	if err := rows.Scan(scanTargets...); err != nil {
		return nil, fmt.Errorf("relationRepo.Get1 scan: %w", err)
	}

	row := scanTargetsToMap(cfg.RefFields, scanTargets)

	return row, nil
}

func (r *RelationRepository) List1N(
	ctx context.Context,
	tx *generated.Tx,
	cfg relation.Config1N,
	mainID int,
	q tableutils.TableQuery,
) (any, error) {

	if len(cfg.RefFields) == 0 {
		return nil, fmt.Errorf("relationRepo.List1N: RefFields is empty")
	}

	selectCols, err := buildSelectCols("r", cfg.RefFields)
	if err != nil {
		return nil, err
	}

	// Build SQL
	baseSQL := fmt.Sprintf(`
        SELECT %s
        FROM %s r
        WHERE r.%s = $1
    `, selectCols, cfg.RefTable, cfg.FKCol)

	orderSQL := tableutils.BuildOrderSQL(q)
	limitSQL := tableutils.BuildLimitSQL(q)

	finalSQL := baseSQL + " " + orderSQL + " " + limitSQL

	rows, err := tx.QueryContext(ctx, finalSQL, mainID)
	if err != nil {
		return nil, fmt.Errorf("relationRepo.List1N query: %w", err)
	}
	defer rows.Close()

	items := make([]map[string]any, 0, 20)

	for rows.Next() {
		scanTargets := buildScanTargets(len(cfg.RefFields))
		if err := rows.Scan(scanTargets...); err != nil {
			return nil, fmt.Errorf("relationRepo.List1N scan: %w", err)
		}

		row := scanTargetsToMap(cfg.RefFields, scanTargets)
		items = append(items, row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("relationRepo.List1N row error: %w", err)
	}

	// Count
	countSQL := fmt.Sprintf(`
        SELECT COUNT(*)
        FROM %s r
        WHERE r.%s = $1
    `, cfg.RefTable, cfg.FKCol)

	var total int

	countRows, err := tx.QueryContext(ctx, countSQL, mainID)
	if err != nil {
		return nil, fmt.Errorf("relationRepo.List1N count query: %w", err)
	}
	defer countRows.Close()

	if countRows.Next() {
		if err := countRows.Scan(&total); err != nil {
			return nil, fmt.Errorf("relationRepo.List1N count scan: %w", err)
		}
	}

	return struct {
		Items []map[string]any `json:"items"`
		Total int              `json:"total"`
	}{
		Items: items,
		Total: total,
	}, nil
}

func (r *RelationRepository) ListM2M(
	ctx context.Context,
	tx *generated.Tx,
	cfg relation.ConfigM2M,
	mainID int,
	q tableutils.TableQuery,
) (any, error) {

	if cfg.RefList == nil || len(cfg.RefList.RefFields) == 0 {
		return nil, fmt.Errorf("relationRepo.ListM2M: RefFields is empty")
	}

	selectCols, err := buildSelectCols("r", cfg.RefList.RefFields)
	if err != nil {
		return nil, err
	}

	mainTable := cfg.MainTable
	refTable := cfg.RefTable
	mainSing := utils.Singular(mainTable)
	refSing := utils.Singular(refTable)

	m2mTable := fmt.Sprintf("%s_%s", mainSing, refTable)
	leftCol := mainSing + "_id"
	rightCol := refSing + "_id"

	baseSQL := fmt.Sprintf(`
		SELECT %s
		FROM %s r
		JOIN %s m2m ON m2m.%s = r.id
		WHERE m2m.%s = $1
	`, selectCols, refTable, m2mTable, rightCol, leftCol)

	orderSQL := tableutils.BuildOrderSQL(q)
	limitSQL := tableutils.BuildLimitSQL(q)

	finalSQL := baseSQL + " " + orderSQL + " " + limitSQL

	rows, err := tx.QueryContext(ctx, finalSQL, mainID)
	if err != nil {
		return nil, fmt.Errorf("relationRepo.List query: %w", err)
	}
	defer rows.Close()

	items := make([]map[string]any, 0, 20)

	for rows.Next() {

		scanTargets := buildScanTargets(len(cfg.RefList.RefFields))

		// Scan row
		if err := rows.Scan(scanTargets...); err != nil {
			return nil, fmt.Errorf("relationRepo.List scan: %w", err)
		}

		row := scanTargetsToMap(cfg.RefList.RefFields, scanTargets)

		items = append(items, row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("relationRepo.List row error: %w", err)
	}

	countSQL := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM %s r
		JOIN %s m2m ON m2m.%s = r.id
		WHERE m2m.%s = $1
	`, refTable, m2mTable, rightCol, leftCol)

	var total int

	countRows, err := tx.QueryContext(ctx, countSQL, mainID)
	if err != nil {
		return nil, fmt.Errorf("relationRepo.List count query: %w", err)
	}
	defer countRows.Close()

	if countRows.Next() {
		if err := countRows.Scan(&total); err != nil {
			return nil, fmt.Errorf("relationRepo.List count scan: %w", err)
		}
	}

	return struct {
		Items []map[string]any `json:"items"`
		Total int              `json:"total"`
	}{
		Items: items,
		Total: total,
	}, nil
}

func (r *RelationRepository) Search(
	ctx context.Context,
	tx *generated.Tx,
	cfg relation.ConfigSearch,
	sq dbutils.SearchQuery,
) (any, error) {

	if len(cfg.RefFields) == 0 {
		return nil, fmt.Errorf("relationRepo.Search: RefFields is empty")
	}

	alias := cfg.Alias
	if alias == "" {
		alias = "r"
	}

	selectCols := ""
	if len(cfg.SelectFields) > 0 {
		if len(cfg.SelectFields) != len(cfg.RefFields) {
			return nil, fmt.Errorf("relationRepo.Search: SelectFields and RefFields length mismatch")
		}
		parts := make([]string, len(cfg.SelectFields))
		for i, expr := range cfg.SelectFields {
			parts[i] = fmt.Sprintf("%s AS %s", expr, cfg.RefFields[i])
		}
		selectCols = strings.Join(parts, ", ")
	} else {
		var err error
		selectCols, err = buildSelectCols(alias, cfg.RefFields)
		if err != nil {
			return nil, err
		}
	}

	refTable := cfg.RefTable

	// BUILD WHERE
	args := make([]any, 0, len(sq.ExtendWhere)+1)
	whereParts := []string{}
	if len(sq.ExtendWhere) > 0 {
		for _, ew := range sq.ExtendWhere {
			appendExtendWhere(&args, &whereParts, alias, ew)
		}
	}

	norm := utils.NormalizeSearchKeyword(sq.Keyword)
	if norm != "" {
		normWhere := dbutils.BuildLikeNormSQLAlias(norm, alias, cfg.NormFields, &args)
		if normWhere != "" {
			whereParts = append(whereParts, normWhere)
		}
	}
	if cfg.ExtraWhere != nil {
		if w := cfg.ExtraWhere(&args); strings.TrimSpace(w) != "" {
			whereParts = append(whereParts, w)
		}
	}

	whereSQL := ""
	if len(whereParts) > 0 {
		whereSQL = "WHERE " + strings.Join(whereParts, " AND ")
	}

	logger.Debug(fmt.Sprintf("[BBB] %v", whereParts))
	logger.Debug(fmt.Sprintf("[BBB-1] %s", whereSQL))

	// BUILD JOINS
	joins := ""
	if cfg.ExtraJoins != nil {
		joins = cfg.ExtraJoins()
	}

	// ORDER BY
	orderField := dbutils.ResolveOrderField(sq.OrderBy, "id")
	direction := "ASC"
	if strings.EqualFold(sq.Direction, "desc") {
		direction = "DESC"
	}
	orderExpr := orderField
	if strings.Contains(orderField, ".") {
		orderExpr = orderField
	} else if len(cfg.SelectFields) == 0 {
		orderExpr = fmt.Sprintf("%s.%s", alias, orderField)
	}
	orderSQL := fmt.Sprintf("ORDER BY %s %s", orderExpr, direction)

	// LIMIT + OFFSET
	limit := sq.Limit
	if limit <= 0 {
		limit = 20
	}
	offset := sq.Offset
	limitSQL := fmt.Sprintf("LIMIT %d OFFSET %d", limit+1, offset)

	// =============================
	// FINAL SQL
	// =============================
	finalSQL := fmt.Sprintf(`
		SELECT %s
		FROM %s %s
		%s
		%s
		%s
		%s
	`, selectCols, refTable, alias, joins, whereSQL, orderSQL, limitSQL)

	logger.Debug(fmt.Sprintf("[AAA] %s", finalSQL))

	rows, err := tx.QueryContext(ctx, finalSQL, args...)
	if err != nil {
		return nil, fmt.Errorf("relationRepo.Search query: %w", err)
	}
	defer rows.Close()

	// SCAN rows
	items := make([]map[string]any, 0, 20)

	for rows.Next() {
		scanTargets := buildScanTargets(len(cfg.RefFields))

		if err := rows.Scan(scanTargets...); err != nil {
			return nil, fmt.Errorf("relationRepo.Search scan: %w", err)
		}

		row := scanTargetsToMap(cfg.RefFields, scanTargets)

		items = append(items, row)
	}

	// Check has_more
	hasMore := false
	totalItems := len(items)
	if totalItems > limit {
		hasMore = true
		items = items[:limit]
	}

	if cfg.OrderRows != nil {
		items = cfg.OrderRows(items)
	}

	// =============================
	// COUNT SQL
	// =============================
	countSQL := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM %s %s
		%s
		%s
	`, refTable, alias, joins, whereSQL)

	countRows, err := tx.QueryContext(ctx, countSQL, args...)
	if err != nil {
		return nil, fmt.Errorf("relationRepo.Search count query: %w", err)
	}
	defer countRows.Close()

	var total int
	if countRows.Next() {
		if err := countRows.Scan(&total); err != nil {
			return nil, fmt.Errorf("relationRepo.Search count scan: %w", err)
		}
	}

	// =============================
	// Convert [] *DTO => [] *any
	// =============================
	n := len(items)
	anyItems := make([]*any, n)

	for i := 0; i < n; i++ {
		tmp := any(items[i])
		anyItems[i] = &tmp
	}

	return dbutils.SearchResult[any]{
		Items:   anyItems,
		HasMore: hasMore,
		Total:   total,
	}, nil
}

func buildSelectCols(prefix string, cols []string) (string, error) {
	if len(cols) == 0 {
		return "", fmt.Errorf("buildSelectCols: empty columns")
	}

	parts := make([]string, len(cols))
	for i, col := range cols {
		if prefix != "" {
			parts[i] = fmt.Sprintf("%s.%s AS %s", prefix, col, col)
		} else {
			parts[i] = col
		}
	}

	return strings.Join(parts, ", "), nil
}

func buildScanTargets(n int) []any {
	targets := make([]any, n)
	for i := 0; i < n; i++ {
		targets[i] = new(any)
	}
	return targets
}

func scanTargetsToMap(cols []string, scanTargets []any) map[string]any {
	row := make(map[string]any, len(cols))

	for i, col := range cols {
		valPtr, ok := scanTargets[i].(*any)
		if !ok || valPtr == nil {
			continue
		}
		row[col] = decodeValue(*valPtr)
	}

	return row
}

func decodeValue(v any) any {
	switch val := v.(type) {
	case json.RawMessage:
		if len(val) == 0 {
			return nil
		}
		var out any
		if err := json.Unmarshal(val, &out); err == nil {
			return out
		}
		return string(val)
	case []byte:
		if len(val) == 0 {
			return nil
		}
		var out any
		if err := json.Unmarshal(val, &out); err == nil {
			return out
		}
		return string(val)
	default:
		return val
	}
}

func appendExtendWhere(args *[]any, whereParts *[]string, alias string, ew any) {
	switch v := ew.(type) {
	case map[string]any:
		for field, val := range v {
			field = strings.TrimSpace(field)
			if !isSafeWhereField(field) {
				continue
			}
			*args = append(*args, val)
			*whereParts = append(*whereParts, fmt.Sprintf("%s = $%d", qualifyWhereField(alias, field), len(*args)))
		}
	case string:
		field, val, ok := parseFieldValue(v)
		if ok && isSafeWhereField(field) {
			*args = append(*args, val)
			*whereParts = append(*whereParts, fmt.Sprintf("%s = $%d", qualifyWhereField(alias, field), len(*args)))
			return
		}
		if strings.TrimSpace(v) != "" {
			*args = append(*args, v)
		}
	default:
		*args = append(*args, v)
	}
}

func parseFieldValue(raw string) (string, any, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", nil, false
	}
	field, val, ok := strings.Cut(raw, "=")
	if !ok {
		return "", nil, false
	}
	field = strings.TrimSpace(field)
	if field == "" {
		return "", nil, false
	}
	return field, strings.TrimSpace(val), true
}

func qualifyWhereField(alias, field string) string {
	if alias == "" || strings.Contains(field, ".") {
		return field
	}
	return fmt.Sprintf("%s.%s", alias, field)
}

func isSafeWhereField(field string) bool {
	if field == "" {
		return false
	}
	prevDot := false
	for i := 0; i < len(field); i++ {
		c := field[i]
		switch {
		case c == '.':
			if i == 0 || i == len(field)-1 || prevDot {
				return false
			}
			prevDot = true
		case c == '_' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9'):
			prevDot = false
		default:
			return false
		}
	}
	return true
}
