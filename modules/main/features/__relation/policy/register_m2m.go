package relation

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/khiemnd777/andy_api/shared/cache"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/logger"
	"github.com/khiemnd777/andy_api/shared/utils"
	"github.com/lib/pq"
)

var (
	mu                       sync.RWMutex
	registry                 = map[string]ConfigM2M{}
	displayOrderColumnExists sync.Map
)

func RegisterM2M(key string, cfg ConfigM2M) {
	if cfg.MainTable == "" || cfg.RefTable == "" {
		panic("relation.Register: missing MainTable or RefTable")
	}
	if cfg.EntityPropMainID == "" || cfg.DTOPropRefIDs == "" {
		panic("relation.Register: missing MainIDProp or RefIDsProp")
	}
	mu.Lock()
	defer mu.Unlock()
	if _, ok := registry[key]; ok {
		panic("relation.Register: duplicate key " + key)
	}
	registry[key] = cfg
}

func GetConfigM2M(key string) (ConfigM2M, error) {
	mu.RLock()
	defer mu.RUnlock()
	cfg, ok := registry[key]
	if !ok {
		return ConfigM2M{}, fmt.Errorf("relation '%s' not registered", key)
	}
	return cfg, nil
}

func UpsertM2M(
	ctx context.Context,
	tx *generated.Tx,
	key string,
	entity any,
	input any,
	output any,
) ([]string, error) {
	cfg, err := GetConfigM2M(key)
	if err != nil {
		return nil, nil
	}

	logger.Debug(fmt.Sprintf("[REL] %v", cfg))

	mainID, err := extractIntField(entity, cfg.EntityPropMainID)
	logger.Debug(fmt.Sprintf("[REL] MainID: %d", mainID))
	if err != nil {
		return nil, fmt.Errorf("relation.Upsert(%s): get main id: %w", key, err)
	}

	ids, err := extractIntSlice(input, cfg.DTOPropRefIDs)
	logger.Debug(fmt.Sprintf("[REL] IDs: %v", ids))
	if err != nil {
		return nil, fmt.Errorf("relation.Upsert(%s): get ids: %w", key, err)
	}

	if ids == nil {
		return nil, nil
	}

	// Dedup + bỏ -1 (ExcludeID mặc định)
	ids = utils.DedupInt(ids, -1)

	var existingRefIDs []int

	extraCols := make([]string, 0, len(cfg.ExtraFields))
	extraVals := make([]any, 0, len(cfg.ExtraFields))
	for _, ef := range cfg.ExtraFields {
		if ef.Column == "" || ef.EntityProp == "" {
			return nil, fmt.Errorf("relation.Upsert(%s): missing column or entity prop in extra field", key)
		}

		val, err := extractFieldInterface(entity, ef.EntityProp)
		if err != nil {
			return nil, fmt.Errorf("relation.Upsert(%s): extract extra field %s: %w", key, ef.EntityProp, err)
		}

		extraCols = append(extraCols, ef.Column)
		extraVals = append(extraVals, val)
	}

	mainTable := cfg.MainTable
	refTable := cfg.RefTable
	mainIDCol := "id"
	refIDCol := "id"
	refNameCol := "name"

	mainSing := utils.Singular(mainTable)
	refSing := utils.Singular(refTable)

	mainNamesCol := refSing + "_names"

	hasRefNameColumn := cfg.RefNameColumn != ""

	var refNames map[int]string
	if hasRefNameColumn && len(ids) > 0 {
		refNames, err = fetchRefNames(ctx, tx, refTable, refIDCol, refNameCol, ids)
		if err != nil {
			return nil, fmt.Errorf("relation.Upsert(%s): fetch ref names: %w", key, err)
		}
	}

	m2mTable := fmt.Sprintf("%s_%s", mainSing, refTable) // "material_suppliers"
	leftCol := mainSing + "_id"                          // "material_id"
	rightCol := refSing + "_id"                          // "supplier_id"

	if cfg.RefValueCache != nil {
		existingRefIDs, err = fetchExistingRefIDs(ctx, tx, m2mTable, leftCol, rightCol, mainID)
		if err != nil {
			return nil, fmt.Errorf("relation.Upsert(%s): fetch existing refs: %w", key, err)
		}
	}

	// 1) Xoá mapping cũ
	delSQL := fmt.Sprintf(`DELETE FROM %s WHERE %s = $1`, m2mTable, leftCol)
	if _, err := tx.ExecContext(ctx, delSQL, mainID); err != nil {
		return nil, fmt.Errorf("relation.Upsert(%s): delete from %s: %w", key, m2mTable, err)
	}

	// 2) Insert lại nếu còn id
	if len(ids) > 0 {
		hasDisplayOrder, err := hasDisplayOrderColumn(ctx, tx, m2mTable)
		if err != nil {
			return nil, fmt.Errorf("relation.Upsert(%s): check display_order column: %w", key, err)
		}

		argPerRow := 2 + len(extraVals)
		if hasDisplayOrder {
			argPerRow++
		}
		if hasRefNameColumn {
			argPerRow++
		}

		vals := make([]string, 0, len(ids))
		args := make([]any, 0, len(ids)*argPerRow)

		for i, id := range ids {
			start := argPerRow*i + 1
			placeholders := make([]string, argPerRow)
			for j := 0; j < argPerRow; j++ {
				placeholders[j] = fmt.Sprintf("$%d", start+j)
			}

			vals = append(vals, fmt.Sprintf("(%s,NOW())", strings.Join(placeholders, ",")))

			rowArgs := []any{mainID, id}
			if hasDisplayOrder {
				rowArgs = append(rowArgs, i)
			}
			rowArgs = append(rowArgs, extraVals...)
			if hasRefNameColumn {
				rowArgs = append(rowArgs, refNames[id])
			}

			args = append(args, rowArgs...)
		}

		columns := []string{leftCol, rightCol}
		if hasDisplayOrder {
			columns = append(columns, "display_order")
		}
		if len(extraCols) > 0 {
			columns = append(columns, extraCols...)
		}
		if hasRefNameColumn {
			columns = append(columns, cfg.RefNameColumn)
		}
		columns = append(columns, "created_at")

		insSQL := fmt.Sprintf(
			`INSERT INTO %s (%s) VALUES %s`,
			m2mTable,
			strings.Join(columns, ","),
			strings.Join(vals, ", "),
		)

		if _, err := tx.ExecContext(ctx, insSQL, args...); err != nil {
			return nil, fmt.Errorf("relation.Upsert(%s): insert into %s: %w", key, m2mTable, err)
		}
	}

	// 3) Lấy danh sách name theo thứ tự ids
	var namesStr string
	names := []string{}

	if len(ids) > 0 {
		updateSQL := fmt.Sprintf(`
			UPDATE %s m
			SET %s = COALESCE((
				SELECT string_agg(r.%s, '|' ORDER BY t.ord)
				FROM unnest($2::int[]) WITH ORDINALITY AS t(%s,ord)
				JOIN %s r ON r.%s = t.%s
			), '')
			WHERE m.%s = $1
			RETURNING %s
		`,
			mainTable,    // materials
			mainNamesCol, // supplier_names
			refNameCol,   // name
			refIDCol,     // id
			refTable,     // suppliers
			refIDCol,     // id
			refIDCol,     // id
			mainIDCol,    // id
			mainNamesCol, // supplier_names
		)

		rows, err := tx.QueryContext(ctx, updateSQL, mainID, pq.Array(ids))
		if err != nil {
			return nil, fmt.Errorf("relation.Upsert(%s): update+return names: %w", key, err)
		}
		defer rows.Close()

		if rows.Next() {
			if err := rows.Scan(&namesStr); err != nil {
				return nil, fmt.Errorf("relation.Upsert(%s): scan namesStr: %w", key, err)
			}
		}
		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("relation.Upsert(%s): rows error: %w", key, err)
		}
	} else {
		// Không có ids -> set rỗng
		updateSQL := fmt.Sprintf(
			`UPDATE %s SET %s = '' WHERE %s = $1 RETURNING %s`,
			mainTable,
			mainNamesCol,
			mainIDCol,
			mainNamesCol,
		)

		logger.Debug(fmt.Sprintf("[REL] update empty names sql: %s", updateSQL))

		rows, err := tx.QueryContext(ctx, updateSQL, mainID)
		if err != nil {
			return nil, fmt.Errorf("relation.Upsert(%s): update empty names: %w", key, err)
		}
		defer rows.Close()

		if rows.Next() {
			if err := rows.Scan(&namesStr); err != nil {
				return nil, fmt.Errorf("relation.Upsert(%s): scan empty namesStr: %w", key, err)
			}
		}
		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("relation.Upsert(%s): rows error (empty): %w", key, err)
		}
	}

	if namesStr != "" {
		names = strings.Split(namesStr, "|")
	}

	// 5) Set result to output
	if cfg.DTOPropDisplayNames != "" {
		if err := setDisplayField(output, cfg.DTOPropDisplayNames, namesStr); err != nil {
			return nil, fmt.Errorf("relation.Upsert(%s): set display value: %w", key, err)
		}
	}

	if cfg.RefValueCache != nil {
		refIDs := utils.DedupInt(append(existingRefIDs, ids...), -1)
		if len(refIDs) > 0 {
			if err := updateRefValueCacheColumns(ctx, tx, cfg, m2mTable, rightCol, refIDCol, refTable, refIDs); err != nil {
				return nil, fmt.Errorf("relation.Upsert(%s): update ref cache columns: %w", key, err)
			}
		}
	}

	// 6) Invalidate
	if cfg.RefList != nil {
		if cfg.RefList.CachePrefix != "" {
			cache.InvalidateKeys(fmt.Sprintf(cfg.RefList.CachePrefix+":%s:%d:*", key, mainID))
		}
	}

	return names, nil
}

func fetchRefNames(ctx context.Context, tx *generated.Tx, table, idCol, nameCol string, ids []int) (map[int]string, error) {
	query := fmt.Sprintf(`SELECT %s, %s FROM %s WHERE %s = ANY($1)`, idCol, nameCol, table, idCol)

	rows, err := tx.QueryContext(ctx, query, pq.Array(ids))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	names := make(map[int]string, len(ids))
	for rows.Next() {
		var (
			id   int
			name sql.NullString
		)
		if err := rows.Scan(&id, &name); err != nil {
			return nil, err
		}
		if name.Valid {
			names[id] = name.String
		} else {
			names[id] = ""
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return names, nil
}

func hasDisplayOrderColumn(ctx context.Context, tx *generated.Tx, table string) (bool, error) {
	if val, ok := displayOrderColumnExists.Load(table); ok {
		return val.(bool), nil
	}

	rows, err := tx.QueryContext(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM information_schema.columns
			WHERE table_schema = current_schema()
			AND table_name = $1
			AND column_name = 'display_order'
		)
	`, table)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	var exists bool
	if rows.Next() {
		if err := rows.Scan(&exists); err != nil {
			return false, err
		}
	} else {
		return false, fmt.Errorf("no rows returned for table %s", table)
	}

	if err := rows.Err(); err != nil {
		return false, err
	}

	displayOrderColumnExists.Store(table, exists)
	return exists, nil
}

func fetchExistingRefIDs(ctx context.Context, tx *generated.Tx, table, leftCol, rightCol string, mainID int) ([]int, error) {
	query := fmt.Sprintf(`SELECT %s FROM %s WHERE %s = $1`, rightCol, table, leftCol)

	rows, err := tx.QueryContext(ctx, query, mainID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return ids, nil
}

func updateRefValueCacheColumns(
	ctx context.Context,
	tx *generated.Tx,
	cfg ConfigM2M,
	m2mTable string,
	rightCol string,
	refIDCol string,
	refTable string,
	refIDs []int,
) error {
	cacheCfg := cfg.RefValueCache
	if cacheCfg == nil || len(cacheCfg.Columns) == 0 || len(refIDs) == 0 {
		return nil
	}

	setClauses := make([]string, 0, len(cacheCfg.Columns))
	selectCols := make([]string, 0, len(cacheCfg.Columns))
	for _, col := range cacheCfg.Columns {
		if col.RefColumn == "" || col.M2MColumn == "" {
			return fmt.Errorf("ref cache column mapping is missing ref or m2m column")
		}
		setClauses = append(setClauses, fmt.Sprintf("%s = latest.%s", col.RefColumn, col.RefColumn))
		selectCols = append(selectCols, fmt.Sprintf("m2m.%s AS %s", col.M2MColumn, col.RefColumn))
	}

	for _, id := range refIDs {
		updateSQL := fmt.Sprintf(`
			UPDATE %s r
			SET %s
			FROM (
				SELECT %s
				FROM %s m2m
				WHERE m2m.%s = %d
				ORDER BY m2m.created_at DESC
				LIMIT 1
			) latest
			WHERE r.%s = %d
		`, refTable, strings.Join(setClauses, ","), strings.Join(selectCols, ","), m2mTable, rightCol, id, refIDCol, id)

		if _, err := tx.ExecContext(ctx, updateSQL); err != nil {
			return fmt.Errorf("update ref cache id %d: %w", id, err)
		}
	}

	return nil
}

// -- helpers
func normalizeStruct(v any) (reflect.Value, error) {
	val := reflect.ValueOf(v)
	if !val.IsValid() {
		return reflect.Value{}, fmt.Errorf("value is nil")
	}
	for val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return reflect.Value{}, fmt.Errorf("value is nil pointer")
		}
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return reflect.Value{}, fmt.Errorf("value is not struct")
	}
	return val, nil
}

func extractIntField(obj any, field string) (int, error) {
	val, err := normalizeStruct(obj)
	if err != nil {
		return 0, err
	}

	f := val.FieldByName(field)
	if !f.IsValid() {
		return 0, fmt.Errorf("field %s not found", field)
	}

	for f.Kind() == reflect.Ptr {
		if f.IsNil() {
			return 0, fmt.Errorf("field %s is nil", field)
		}
		f = f.Elem()
	}

	switch f.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return int(f.Int()), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return int(f.Uint()), nil
	default:
		return 0, fmt.Errorf("field %s is not int/uint", field)
	}
}

func extractIntSlice(obj any, field string) ([]int, error) {
	val, err := normalizeStruct(obj)
	if err != nil {
		return nil, err
	}

	f := val.FieldByName(field)
	if !f.IsValid() {
		return nil, fmt.Errorf("field %s not found", field)
	}

	for f.Kind() == reflect.Ptr {
		if f.IsNil() {
			return nil, nil
		}
		f = f.Elem()
	}

	if f.Kind() != reflect.Slice {
		return nil, fmt.Errorf("field %s is not slice", field)
	}

	if f.IsNil() {
		return nil, nil
	}

	n := f.Len()
	out := make([]int, n)
	for i := 0; i < n; i++ {
		el := f.Index(i)
		for el.Kind() == reflect.Ptr {
			if el.IsNil() {
				return nil, fmt.Errorf("field %s slice element is nil pointer", field)
			}
			el = el.Elem()
		}
		switch el.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			out[i] = int(el.Int())
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			out[i] = int(el.Uint())
		default:
			return nil, fmt.Errorf("field %s slice element is not int/uint", field)
		}
	}

	return out, nil
}

func extractFieldInterface(obj any, field string) (any, error) {
	val, err := normalizeStruct(obj)
	if err != nil {
		return nil, err
	}

	f := val.FieldByName(field)
	if !f.IsValid() {
		return nil, fmt.Errorf("field %s not found", field)
	}

	for f.Kind() == reflect.Ptr {
		if f.IsNil() {
			return nil, nil
		}
		f = f.Elem()
	}

	return f.Interface(), nil
}

func setDisplayField(obj any, field string, val string) error {
	target, err := normalizeStruct(obj)
	if err != nil {
		return err
	}

	f := target.FieldByName(field)
	if !f.IsValid() {
		return fmt.Errorf("field %s not found", field)
	}
	if !f.CanSet() {
		return fmt.Errorf("field %s cannot be set", field)
	}

	switch f.Kind() {
	case reflect.String:
		f.SetString(val)
		return nil
	case reflect.Ptr:
		if f.Type().Elem().Kind() != reflect.String {
			return fmt.Errorf("field %s is not *string", field)
		}
		f.Set(reflect.ValueOf(&val))
		return nil
	default:
		return fmt.Errorf("field %s must be string or *string", field)
	}
}
