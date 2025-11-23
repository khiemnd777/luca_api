package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	relation "github.com/khiemnd777/andy_api/modules/main/features/__relation/policy"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	dbutils "github.com/khiemnd777/andy_api/shared/db/utils"
	"github.com/khiemnd777/andy_api/shared/utils"
	tableutils "github.com/khiemnd777/andy_api/shared/utils/table"
)

type RelationRepository struct{}

func NewRelationRepository() *RelationRepository {
	return &RelationRepository{}
}

func (r *RelationRepository) List(
	ctx context.Context,
	tx *generated.Tx,
	cfg relation.Config,
	mainID int,
	q tableutils.TableQuery,
) (any, error) {

	dtoType := reflect.TypeOf(cfg.GetRefList.RefDTO)
	if dtoType.Kind() == reflect.Ptr {
		dtoType = dtoType.Elem()
	}

	sliceType := reflect.SliceOf(reflect.PointerTo(dtoType))
	sliceValue := reflect.MakeSlice(sliceType, 0, 20)

	cols := make([]string, 0, dtoType.NumField())

	for i := 0; i < dtoType.NumField(); i++ {
		f := dtoType.Field(i)
		colName := utils.ToSnake(f.Name)
		cols = append(cols, "r."+colName)
	}

	selectCols := strings.Join(cols, ", ")

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

	for rows.Next() {

		elem := reflect.New(dtoType)
		elemVal := elem.Elem()

		scanTargets := make([]any, dtoType.NumField())

		// build pointer list
		for i := 0; i < dtoType.NumField(); i++ {
			f := elemVal.Field(i)

			// JSONB → scan vào []byte trước
			if f.Kind() == reflect.Map && f.Type().String() == "map[string]interface {}" {
				var raw json.RawMessage
				scanTargets[i] = &raw
			} else {
				scanTargets[i] = f.Addr().Interface()
			}
		}

		// Scan row
		if err := rows.Scan(scanTargets...); err != nil {
			return nil, fmt.Errorf("relationRepo.List scan: %w", err)
		}

		// Convert JSONB → map[string]any
		for i := 0; i < dtoType.NumField(); i++ {
			f := elemVal.Field(i)

			if f.Kind() == reflect.Map && f.Type().String() == "map[string]interface {}" {
				rawPtr, ok := scanTargets[i].(*json.RawMessage)
				if !ok || rawPtr == nil {
					continue
				}
				if len(*rawPtr) == 0 {
					continue
				}

				var m map[string]any
				if e := json.Unmarshal(*rawPtr, &m); e == nil {
					f.Set(reflect.ValueOf(m))
				}
			}
		}

		sliceValue = reflect.Append(sliceValue, elem)
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

	ptrDtoType := reflect.PointerTo(dtoType)

	tableListType := reflect.StructOf([]reflect.StructField{
		{
			Name: "Items",
			Type: reflect.SliceOf(ptrDtoType),
			Tag:  `json:"items"`,
		},
		{
			Name: "Total",
			Type: reflect.TypeOf(int(0)),
			Tag:  `json:"total"`,
		},
	})

	out := reflect.New(tableListType).Elem()
	out.FieldByName("Items").Set(sliceValue)
	out.FieldByName("Total").SetInt(int64(total))

	return out.Interface(), nil
}

func (r *RelationRepository) Search(
	ctx context.Context,
	tx *generated.Tx,
	cfg relation.Config,
	sq dbutils.SearchQuery,
) (any, error) {

	dtoType := reflect.TypeOf(cfg.GetRefSearch.RefDTO)
	if dtoType.Kind() == reflect.Ptr {
		dtoType = dtoType.Elem()
	}

	// build SELECT columns
	cols := make([]string, 0, dtoType.NumField())
	for i := 0; i < dtoType.NumField(); i++ {
		f := dtoType.Field(i)
		colName := utils.ToSnake(f.Name)
		cols = append(cols, "r."+colName)
	}
	selectCols := strings.Join(cols, ", ")

	refTable := cfg.RefTable

	// =============================
	// BUILD WHERE
	// =============================
	args := []any{}
	whereParts := []string{}

	norm := utils.NormalizeSearchKeyword(sq.Keyword)
	if norm != "" {
		likeParts := make([]string, 0)
		for i := 0; i < dtoType.NumField(); i++ {
			f := dtoType.Field(i)

			// allow string or *string
			isString :=
				f.Type.Kind() == reflect.String ||
					(f.Type.Kind() == reflect.Ptr && f.Type.Elem().Kind() == reflect.String)

			if !isString {
				continue
			}

			col := utils.ToSnake(f.Name)
			likeParts = append(likeParts,
				fmt.Sprintf("LOWER(r.%s) LIKE $%d", col, len(args)+1))
			args = append(args, "%"+norm+"%")
		}

		whereParts = append(whereParts, "("+strings.Join(likeParts, " OR ")+")")
	}

	whereSQL := ""
	if len(whereParts) > 0 {
		whereSQL = "WHERE " + strings.Join(whereParts, " AND ")
	}

	// =============================
	// ORDER BY
	// =============================
	orderField := dbutils.ResolveOrderField(sq.OrderBy, "id")
	direction := "ASC"
	if strings.EqualFold(sq.Direction, "desc") {
		direction = "DESC"
	}
	orderSQL := fmt.Sprintf("ORDER BY r.%s %s", orderField, direction)

	// =============================
	// LIMIT + OFFSET
	// =============================
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
		FROM %s r
		%s
		%s
		%s
	`, selectCols, refTable, whereSQL, orderSQL, limitSQL)

	rows, err := tx.QueryContext(ctx, finalSQL, args...)
	if err != nil {
		return nil, fmt.Errorf("relationRepo.Search query: %w", err)
	}
	defer rows.Close()

	// =============================
	// SCAN rows
	// =============================
	ptrType := reflect.PointerTo(dtoType)
	sliceType := reflect.SliceOf(ptrType)
	sliceValue := reflect.MakeSlice(sliceType, 0, 20)

	for rows.Next() {
		elem := reflect.New(dtoType)
		elemVal := elem.Elem()

		scanTargets := make([]any, dtoType.NumField())

		for i := 0; i < dtoType.NumField(); i++ {
			f := elemVal.Field(i)
			if f.Kind() == reflect.Map && f.Type().String() == "map[string]interface {}" {
				var raw json.RawMessage
				scanTargets[i] = &raw
			} else {
				scanTargets[i] = f.Addr().Interface()
			}
		}

		if err := rows.Scan(scanTargets...); err != nil {
			return nil, fmt.Errorf("relationRepo.Search scan: %w", err)
		}

		// convert JSONB
		for i := 0; i < dtoType.NumField(); i++ {
			f := elemVal.Field(i)
			if f.Kind() == reflect.Map && f.Type().String() == "map[string]interface {}" {
				raw, ok := scanTargets[i].(*json.RawMessage)
				if ok && raw != nil && len(*raw) > 0 {
					var m map[string]any
					json.Unmarshal(*raw, &m)
					f.Set(reflect.ValueOf(m))
				}
			}
		}

		sliceValue = reflect.Append(sliceValue, elem)
	}

	// =============================
	// Check has_more
	// =============================
	hasMore := false
	totalItems := sliceValue.Len()
	if totalItems > limit {
		hasMore = true
		sliceValue = sliceValue.Slice(0, limit)
	}

	// =============================
	// COUNT SQL
	// =============================
	countSQL := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM %s r
		%s
	`, refTable, whereSQL)

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
	n := sliceValue.Len()
	items := make([]*any, n)

	for i := 0; i < n; i++ {
		v := sliceValue.Index(i).Interface() // *DTO
		tmp := any(v)
		items[i] = &tmp
	}

	return dbutils.SearchResult[any]{
		Items:   items,
		HasMore: hasMore,
		Total:   total,
	}, nil
}
