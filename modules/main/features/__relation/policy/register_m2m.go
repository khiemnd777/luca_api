package relation

import (
	"context"
	"fmt"
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
	if cfg.GetMainID == nil || cfg.GetIDs == nil {
		panic("relation.Register: missing GetMainID or GetIDs")
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

	mainID, err := cfg.GetMainID(entity)
	logger.Debug(fmt.Sprintf("[REL] MainID: %d", mainID))
	if err != nil {
		return nil, fmt.Errorf("relation.Upsert(%s): GetMainID: %w", key, err)
	}

	ids, err := cfg.GetIDs(input)
	logger.Debug(fmt.Sprintf("[REL] IDs: %v", ids))
	if err != nil {
		return nil, fmt.Errorf("relation.Upsert(%s): GetIDs: %w", key, err)
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
	err = cfg.SetResult(output, ids, &namesStr, names)
	if err != nil {
		return nil, fmt.Errorf("relation.Upsert(%s): SetResult: %w", key, err)
	}

	// 6) Invalidate
	if cfg.GetRefList != nil {
		if cfg.GetRefList.CachePrefix != "" {
			cache.InvalidateKeys(fmt.Sprintf(cfg.GetRefList.CachePrefix+":%s:%d:*", key, mainID))
		}
	}

	return names, nil
}
