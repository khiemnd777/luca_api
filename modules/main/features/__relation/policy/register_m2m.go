package relation

import (
	"context"
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
	mu       sync.RWMutex
	registry = map[string]ConfigM2M{}
)

func RegisterM2M(key string, cfg ConfigM2M) {
	if cfg.MainTable == "" || cfg.RefTable == "" {
		panic("relation.Register: missing MainTable or RefTable")
	}
	if cfg.MainIDProp == "" || cfg.RefIDsProp == "" || cfg.DisplayProp == "" {
		panic("relation.Register: missing MainIDProp or RefIDsProp or DisplayProp")
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

	mainID, err := extractIntField(entity, cfg.MainIDProp)
	logger.Debug(fmt.Sprintf("[REL] MainID: %d", mainID))
	if err != nil {
		return nil, fmt.Errorf("relation.Upsert(%s): get main id: %w", key, err)
	}

	ids, err := extractIntSlice(input, cfg.RefIDsProp)
	logger.Debug(fmt.Sprintf("[REL] IDs: %v", ids))
	if err != nil {
		return nil, fmt.Errorf("relation.Upsert(%s): get ids: %w", key, err)
	}

	if ids == nil {
		return nil, nil
	}

	// Dedup + bỏ -1 (ExcludeID mặc định)
	ids = utils.DedupInt(ids, -1)

	mainTable := cfg.MainTable
	refTable := cfg.RefTable
	mainIDCol := "id"
	refIDCol := "id"
	refNameCol := "name"

	mainSing := utils.Singular(mainTable)
	refSing := utils.Singular(refTable)

	mainNamesCol := refSing + "_names"

	m2mTable := fmt.Sprintf("%s_%s", mainSing, refTable) // "material_suppliers"
	leftCol := mainSing + "_id"                          // "material_id"
	rightCol := refSing + "_id"                          // "supplier_id"

	// 1) Xoá mapping cũ
	delSQL := fmt.Sprintf(`DELETE FROM %s WHERE %s = $1`, m2mTable, leftCol)
	if _, err := tx.ExecContext(ctx, delSQL, mainID); err != nil {
		return nil, fmt.Errorf("relation.Upsert(%s): delete from %s: %w", key, m2mTable, err)
	}

	// 2) Insert lại nếu còn id
	if len(ids) > 0 {
		vals := make([]string, 0, len(ids))
		args := make([]any, 0, len(ids)*2)

		for i, id := range ids {
			p1 := 2*i + 1
			p2 := 2*i + 2

			// thêm NOW() vào từng row
			vals = append(vals, fmt.Sprintf("($%d,$%d,NOW())", p1, p2))

			args = append(args, mainID, id)
		}

		insSQL := fmt.Sprintf(
			`INSERT INTO %s (%s,%s,created_at) VALUES %s`,
			m2mTable,
			leftCol,
			rightCol,
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

		logger.Debug(fmt.Sprintf("[REL] update+names sql: %s", updateSQL))

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
	if err := setDisplayField(output, cfg.DisplayProp, namesStr); err != nil {
		return nil, fmt.Errorf("relation.Upsert(%s): set display value: %w", key, err)
	}

	// 6) Invalidate
	if cfg.RefList != nil {
		if cfg.RefList.CachePrefix != "" {
			cache.InvalidateKeys(fmt.Sprintf(cfg.RefList.CachePrefix+":%s:%d:*", key, mainID))
		}
	}

	return names, nil
}

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
