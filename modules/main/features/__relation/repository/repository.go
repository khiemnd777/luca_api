package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	relation "github.com/khiemnd777/andy_api/modules/main/features/__relation/policy"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
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
