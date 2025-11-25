package relation

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
)

var (
	mu1       sync.RWMutex
	registry1 = map[string]Config1{}
)

func Register1(key string, cfg Config1) {
	mu1.Lock()
	defer mu1.Unlock()
	if _, ok := registry1[key]; ok {
		panic("relation.Register: duplicate key " + key)
	}
	registry1[key] = cfg
}

func GetConfig1(key string) (Config1, error) {
	mu1.RLock()
	defer mu1.RUnlock()
	cfg, ok := registry1[key]
	if !ok {
		return Config1{}, fmt.Errorf("relation '%s' not registered", key)
	}
	return cfg, nil
}

func Upsert1(
	ctx context.Context,
	tx *generated.Tx,
	key string,
	entity any,
	inputDTO any,
	outputDTO any,
) error {
	cfg, err := GetConfig1(key)
	if err != nil {
		return nil
	}

	mainVal := reflect.ValueOf(entity).Elem()
	mainID := int(mainVal.FieldByName(cfg.MainIDProp).Int())

	inV := reflect.ValueOf(inputDTO)
	if inV.Kind() != reflect.Ptr {
		return fmt.Errorf("Upsert1: inputDTO must be a pointer, got %s", inV.Kind())
	}
	inVal := inV.Elem()
	refID := int(inVal.FieldByName(cfg.UpsertedIDProp).Int())

	if refID <= 0 {

		// Build dynamic SET clause
		setParts := []string{fmt.Sprintf("%s = NULL", cfg.MainRefIDCol)}

		if cfg.MainRefNameCol != nil {
			setParts = append(setParts, fmt.Sprintf("%s = NULL", *cfg.MainRefNameCol))
		}

		returnCols := []string{cfg.MainRefIDCol}
		if cfg.MainRefNameCol != nil {
			returnCols = append(returnCols, *cfg.MainRefNameCol)
		}

		sql := fmt.Sprintf(`
            UPDATE %s
            SET %s
            WHERE %s = $1
            RETURNING %s
        `,
			cfg.MainTable,
			strings.Join(setParts, ", "),
			cfg.MainIDProp,
			strings.Join(returnCols, ", "),
		)

		rows, err := tx.QueryContext(ctx, sql, mainID)
		if err != nil {
			return fmt.Errorf("Upsert1 clear: %w", err)
		}
		defer rows.Close()

		outVal := reflect.ValueOf(outputDTO).Elem()
		outVal.FieldByName(cfg.UpsertedIDProp).SetInt(0)

		if cfg.UpsertedNameProp != nil {
			outVal.FieldByName(*cfg.UpsertedNameProp).SetString("")
		}

		return nil
	}

	setParts := []string{
		fmt.Sprintf("%s = $2", cfg.MainRefIDCol),
	}

	nameSelect := ""
	returnCols := []string{cfg.MainRefIDCol}

	if cfg.MainRefNameCol != nil {
		nameSelect = fmt.Sprintf(
			"(SELECT %s FROM %s WHERE %s = $2)",
			cfg.RefNameCol,
			cfg.RefTable,
			cfg.RefIDCol,
		)

		setParts = append(setParts,
			fmt.Sprintf("%s = %s", *cfg.MainRefNameCol, nameSelect),
		)

		returnCols = append(returnCols, *cfg.MainRefNameCol)
	}

	updateSQL := fmt.Sprintf(`
        UPDATE %s AS m
        SET %s
        WHERE m.%s = $1
        RETURNING %s
    `,
		cfg.MainTable,
		strings.Join(setParts, ", "),
		cfg.MainIDProp,
		strings.Join(returnCols, ", "),
	)

	rows, err := tx.QueryContext(ctx, updateSQL, mainID, refID)
	if err != nil {
		return fmt.Errorf("Upsert1 update: %w", err)
	}
	defer rows.Close()

	var outRefID int
	var outName *string

	if cfg.MainRefNameCol != nil {
		// expecting: refID, refName
		if rows.Next() {
			if err := rows.Scan(&outRefID, &outName); err != nil {
				return fmt.Errorf("Upsert1 scan: %w", err)
			}
		}
	} else {
		// expecting: only refID
		if rows.Next() {
			if err := rows.Scan(&outRefID); err != nil {
				return fmt.Errorf("Upsert1 scan (id only): %w", err)
			}
		}
	}

	finalName := ""
	if outName != nil {
		finalName = *outName
	}

	outVal := reflect.ValueOf(outputDTO).Elem()

	outVal.FieldByName(cfg.UpsertedIDProp).SetInt(int64(outRefID))

	if cfg.UpsertedNameProp != nil {
		fv := outVal.FieldByName(*cfg.UpsertedNameProp)
		if !fv.IsValid() {
			return fmt.Errorf("Upsert1: field %s not found", *cfg.UpsertedNameProp)
		}

		switch fv.Kind() {
		case reflect.Ptr:
			if finalName == "" {
				fv.Set(reflect.Zero(fv.Type())) // nil
			} else {
				fv.Set(reflect.ValueOf(&finalName))
			}

		case reflect.String:
			fv.SetString(finalName)

		default:
			return fmt.Errorf("Upsert1: cannot set string on field %s of kind %s",
				*cfg.UpsertedNameProp,
				fv.Kind(),
			)
		}
	}

	return nil
}
