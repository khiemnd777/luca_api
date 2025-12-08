package relation

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/khiemnd777/andy_api/shared/cache"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
)

type idxItem struct {
	idx  int
	item reflect.Value
}

func Upsert1N(ctx context.Context, tx *generated.Tx, key string, parentID int, rawItems any) (any, error) {
	cfg, err := GetConfig1N(key)
	if err != nil {
		return nil, err
	}

	if err := validateConfig1N(cfg); err != nil {
		return nil, err
	}

	itemsVal, err := normalizeSliceValue(rawItems)
	if err != nil {
		return nil, err
	}

	out := reflect.MakeSlice(itemsVal.Type(), itemsVal.Len(), itemsVal.Len())

	if itemsVal.Len() == 0 {
		return out.Interface(), nil
	}

	newItems := make([]idxItem, 0, itemsVal.Len())
	updateItems := make([]idxItem, 0, itemsVal.Len())

	for i := 0; i < itemsVal.Len(); i++ {
		item := itemsVal.Index(i)
		normalized, err := ensurePointerStruct(item)
		if err != nil {
			return nil, err
		}

		if err := setIntField(normalized, cfg.ParentIDProp, parentID); err != nil {
			return nil, err
		}

		id, err := extractIntFieldValue(normalized, cfg.IDProp)
		if err != nil {
			return nil, err
		}

		if id > 0 {
			updateItems = append(updateItems, idxItem{idx: i, item: normalized})
		} else {
			newItems = append(newItems, idxItem{idx: i, item: normalized})
		}

		out.Index(i).Set(normalized)
	}

	if len(newItems) > 0 {
		if err := bulkInsert1N(ctx, tx, cfg, parentID, newItems); err != nil {
			return nil, err
		}
	}

	if len(updateItems) > 0 {
		if err := update1N(ctx, tx, cfg, parentID, updateItems); err != nil {
			return nil, err
		}
	}

	if cfg.CachePrefix != "" {
		cache.InvalidateKeys(fmt.Sprintf("%s:%s:%d:*", cfg.CachePrefix, key, parentID))
	}

	return out.Interface(), nil
}

func bulkInsert1N(ctx context.Context, tx *generated.Tx, cfg Config1N, parentID int, items []idxItem) error {
	colNames := append([]string{cfg.FKCol}, cfg.InsertCols...)
	colCount := len(colNames)

	placeholders := make([]string, 0, len(items))
	args := make([]any, 0, len(items)*colCount)

	for i, it := range items {
		start := i*colCount + 1

		ph := make([]string, colCount)
		for j := 0; j < colCount; j++ {
			ph[j] = fmt.Sprintf("$%d", start+j)
		}
		placeholders = append(placeholders, "("+strings.Join(ph, ",")+")")

		args = append(args, parentID)
		for _, prop := range cfg.InsertProps {
			val, err := fieldValueInterface(it.item, prop)
			if err != nil {
				return err
			}
			args = append(args, val)
		}
	}

	sqlStr := fmt.Sprintf(`
		INSERT INTO %s (%s)
		VALUES %s
		RETURNING %s
	`, cfg.RefTable, strings.Join(colNames, ","), strings.Join(placeholders, ","), strings.Join(cfg.ReturnCols, ","))

	rows, err := tx.QueryContext(ctx, sqlStr, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	idx := 0

	for rows.Next() {
		if idx >= len(items) {
			return fmt.Errorf("bulkInsert1N: unexpected extra row")
		}

		scanTargets, setters, err := buildScanPlan(items[idx].item, cfg.ReturnProps)
		if err != nil {
			return err
		}

		if err := rows.Scan(scanTargets...); err != nil {
			return err
		}

		if err := applySetters(setters); err != nil {
			return err
		}

		if err := setIntField(items[idx].item, cfg.ParentIDProp, parentID); err != nil {
			return err
		}

		idx++
	}

	if idx != len(items) {
		return fmt.Errorf("bulkInsert1N: expected %d rows, got %d", len(items), idx)
	}

	return rows.Err()
}

func update1N(ctx context.Context, tx *generated.Tx, cfg Config1N, parentID int, items []idxItem) error {
	setParts := []string{fmt.Sprintf("%s = $1", cfg.FKCol)}
	for i, col := range cfg.InsertCols {
		setParts = append(setParts, fmt.Sprintf("%s = $%d", col, i+2))
	}

	if cfg.UpdatedAtCol != "" {
		setParts = append(setParts, fmt.Sprintf("%s = NOW()", cfg.UpdatedAtCol))
	}

	whereIdx := len(cfg.InsertCols) + 2

	sqlStr := fmt.Sprintf(`
		UPDATE %s
		SET %s
		WHERE %s = $%d
		RETURNING %s
	`, cfg.RefTable, strings.Join(setParts, ", "), cfg.IDCol, whereIdx, strings.Join(cfg.ReturnCols, ","))

	for _, it := range items {
		args := make([]any, 0, len(cfg.InsertCols)+2)
		args = append(args, parentID)

		for _, prop := range cfg.InsertProps {
			val, err := fieldValueInterface(it.item, prop)
			if err != nil {
				return err
			}
			args = append(args, val)
		}

		idVal, err := extractIntFieldValue(it.item, cfg.IDProp)
		if err != nil {
			return err
		}
		args = append(args, idVal)

		scanTargets, setters, err := buildScanPlan(it.item, cfg.ReturnProps)
		if err != nil {
			return err
		}

		rows, err := tx.QueryContext(ctx, sqlStr, args...)
		if err != nil {
			return err
		}

		var scanErr error
		if rows.Next() {
			scanErr = rows.Scan(scanTargets...)
		} else {
			scanErr = sql.ErrNoRows
		}

		closeErr := rows.Close()
		if scanErr != nil {
			return scanErr
		}
		if closeErr != nil {
			return closeErr
		}
		if err := rows.Err(); err != nil {
			return err
		}

		if err := applySetters(setters); err != nil {
			return err
		}

		if err := setIntField(it.item, cfg.ParentIDProp, parentID); err != nil {
			return err
		}
	}

	return nil
}

func validateConfig1N(cfg Config1N) error {
	if cfg.RefTable == "" || cfg.FKCol == "" || cfg.IDCol == "" {
		return fmt.Errorf("relation.Upsert1N: missing table or column config")
	}
	if cfg.IDProp == "" || cfg.ParentIDProp == "" {
		return fmt.Errorf("relation.Upsert1N: missing id or parent prop")
	}
	if len(cfg.InsertCols) != len(cfg.InsertProps) {
		return fmt.Errorf("relation.Upsert1N: InsertCols and InsertProps length mismatch")
	}
	if len(cfg.ReturnCols) != len(cfg.ReturnProps) {
		return fmt.Errorf("relation.Upsert1N: ReturnCols and ReturnProps length mismatch")
	}
	if len(cfg.ReturnCols) == 0 {
		return fmt.Errorf("relation.Upsert1N: ReturnCols cannot be empty")
	}
	return nil
}

func normalizeSliceValue(items any) (reflect.Value, error) {
	val := reflect.ValueOf(items)
	if !val.IsValid() {
		return reflect.Value{}, fmt.Errorf("relation.Upsert1N: items is nil")
	}
	if val.Kind() != reflect.Slice {
		return reflect.Value{}, fmt.Errorf("relation.Upsert1N: items must be slice, got %s", val.Kind())
	}
	return val, nil
}

func ensurePointerStruct(v reflect.Value) (reflect.Value, error) {
	if v.Kind() != reflect.Ptr {
		return reflect.Value{}, fmt.Errorf("relation.Upsert1N: item must be pointer, got %s", v.Kind())
	}
	if v.IsNil() {
		v.Set(reflect.New(v.Type().Elem()))
	}
	if v.Elem().Kind() != reflect.Struct {
		return reflect.Value{}, fmt.Errorf("relation.Upsert1N: item must point to struct, got %s", v.Elem().Kind())
	}
	return v, nil
}

func getStructField(v reflect.Value, name string) (reflect.Value, error) {
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return reflect.Value{}, fmt.Errorf("relation.Upsert1N: item must be non-nil pointer")
	}

	f := v.Elem().FieldByName(name)
	if !f.IsValid() {
		return reflect.Value{}, fmt.Errorf("relation.Upsert1N: field %s not found", name)
	}

	if !f.CanSet() {
		return reflect.Value{}, fmt.Errorf("relation.Upsert1N: field %s cannot be set", name)
	}

	return f, nil
}

func setIntField(v reflect.Value, name string, value int) error {
	f, err := getStructField(v, name)
	if err != nil {
		return err
	}

	switch f.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		f.SetInt(int64(value))
	case reflect.Ptr:
		if f.Type().Elem().Kind() >= reflect.Int && f.Type().Elem().Kind() <= reflect.Int64 {
			val := reflect.New(f.Type().Elem())
			val.Elem().SetInt(int64(value))
			f.Set(val)
			return nil
		}
		return fmt.Errorf("relation.Upsert1N: field %s is not int pointer", name)
	default:
		return fmt.Errorf("relation.Upsert1N: field %s is not int type", name)
	}

	return nil
}

func extractIntFieldValue(v reflect.Value, name string) (int, error) {
	f, err := getStructField(v, name)
	if err != nil {
		return 0, err
	}

	switch f.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return int(f.Int()), nil
	case reflect.Ptr:
		if f.IsNil() {
			return 0, nil
		}
		if f.Type().Elem().Kind() >= reflect.Int && f.Type().Elem().Kind() <= reflect.Int64 {
			return int(f.Elem().Int()), nil
		}
		return 0, fmt.Errorf("relation.Upsert1N: field %s is not int pointer", name)
	default:
		return 0, fmt.Errorf("relation.Upsert1N: field %s is not int type", name)
	}
}

func fieldValueInterface(v reflect.Value, name string) (any, error) {
	f, err := getStructField(v, name)
	if err != nil {
		return nil, err
	}

	if f.Kind() == reflect.Ptr {
		if f.IsNil() {
			return nil, nil
		}
		return f.Interface(), nil
	}

	return f.Interface(), nil
}

func buildScanPlan(v reflect.Value, props []string) ([]any, []func() error, error) {
	targets := make([]any, len(props))
	setters := make([]func() error, len(props))

	for i, prop := range props {
		field, err := getStructField(v, prop)
		if err != nil {
			return nil, nil, err
		}

		switch field.Kind() {
		case reflect.Ptr:
			elemKind := field.Type().Elem().Kind()
			switch elemKind {
			case reflect.String:
				var ns sql.NullString
				targets[i] = &ns
				setters[i] = func(f reflect.Value, val *sql.NullString) func() error {
					return func() error {
						if val.Valid {
							s := val.String
							f.Set(reflect.ValueOf(&s))
						} else {
							f.Set(reflect.Zero(f.Type()))
						}
						return nil
					}
				}(field, &ns)
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				var ni sql.NullInt64
				targets[i] = &ni
				setters[i] = func(f reflect.Value, val *sql.NullInt64) func() error {
					return func() error {
						if val.Valid {
							n := val.Int64
							ptr := reflect.New(f.Type().Elem())
							ptr.Elem().SetInt(n)
							f.Set(ptr)
						} else {
							f.Set(reflect.Zero(f.Type()))
						}
						return nil
					}
				}(field, &ni)
			case reflect.Struct:
				if field.Type().Elem() == reflect.TypeOf(time.Time{}) {
					var t time.Time
					targets[i] = &t
					setters[i] = func(f reflect.Value, val *time.Time) func() error {
						return func() error {
							ptr := reflect.New(f.Type().Elem())
							ptr.Elem().Set(reflect.ValueOf(*val))
							f.Set(ptr)
							return nil
						}
					}(field, &t)
				} else {
					target := reflect.New(field.Type().Elem())
					targets[i] = target.Interface()
					setters[i] = func(f reflect.Value, tgt reflect.Value) func() error {
						return func() error {
							f.Set(tgt)
							return nil
						}
					}(field, target)
				}
			case reflect.Map:
				var raw []byte
				targets[i] = &raw
				setters[i] = buildJSONSetter(field, &raw)
			default:
				target := reflect.New(field.Type().Elem())
				targets[i] = target.Interface()
				setters[i] = func(f reflect.Value, tgt reflect.Value) func() error {
					return func() error {
						f.Set(tgt)
						return nil
					}
				}(field, target)
			}

		case reflect.String:
			var ns sql.NullString
			targets[i] = &ns
			setters[i] = func(f reflect.Value, val *sql.NullString) func() error {
				return func() error {
					if val.Valid {
						f.SetString(val.String)
					} else {
						f.SetString("")
					}
					return nil
				}
			}(field, &ns)

		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			var ni sql.NullInt64
			targets[i] = &ni
			setters[i] = func(f reflect.Value, val *sql.NullInt64) func() error {
				return func() error {
					if val.Valid {
						f.SetInt(val.Int64)
					} else {
						f.SetInt(0)
					}
					return nil
				}
			}(field, &ni)

		case reflect.Map:
			var raw []byte
			targets[i] = &raw
			setters[i] = buildJSONSetter(field, &raw)

		case reflect.Struct:
			if field.Type() == reflect.TypeOf(time.Time{}) {
				var t time.Time
				targets[i] = &t
				setters[i] = func(f reflect.Value, val *time.Time) func() error {
					return func() error {
						f.Set(reflect.ValueOf(*val))
						return nil
					}
				}(field, &t)
			} else {
				target := reflect.New(field.Type())
				targets[i] = target.Interface()
				setters[i] = func(f reflect.Value, tgt reflect.Value) func() error {
					return func() error {
						f.Set(tgt.Elem())
						return nil
					}
				}(field, target)
			}

		default:
			return nil, nil, fmt.Errorf("relation.Upsert1N: unsupported field kind %s for prop %s", field.Kind(), prop)
		}
	}

	return targets, setters, nil
}

func buildJSONSetter(field reflect.Value, raw *[]byte) func() error {
	return func() error {
		if len(*raw) == 0 {
			field.Set(reflect.Zero(field.Type()))
			return nil
		}

		if field.Kind() == reflect.Ptr {
			target := reflect.New(field.Type().Elem())
			if err := json.Unmarshal(*raw, target.Interface()); err != nil {
				return err
			}
			field.Set(target)
			return nil
		}

		target := reflect.New(field.Type())
		if err := json.Unmarshal(*raw, target.Interface()); err != nil {
			return err
		}
		field.Set(target.Elem())

		return nil
	}
}

func applySetters(setters []func() error) error {
	for _, set := range setters {
		if err := set(); err != nil {
			return err
		}
	}
	return nil
}
