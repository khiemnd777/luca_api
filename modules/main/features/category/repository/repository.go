package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"strconv"
	"time"

	"github.com/khiemnd777/andy_api/modules/main/config"
	model "github.com/khiemnd777/andy_api/modules/main/features/__model"
	relation "github.com/khiemnd777/andy_api/modules/main/features/__relation/policy"
	"github.com/khiemnd777/andy_api/modules/metadata/repository"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/category"
	dbutils "github.com/khiemnd777/andy_api/shared/db/utils"
	"github.com/khiemnd777/andy_api/shared/mapper"
	"github.com/khiemnd777/andy_api/shared/metadata/customfields"
	"github.com/khiemnd777/andy_api/shared/module"
	"github.com/khiemnd777/andy_api/shared/utils/table"
)

type CategoryRepository interface {
	Create(ctx context.Context, input *model.CategoryUpsertDTO) (*model.CategoryDTO, error)
	Update(ctx context.Context, input *model.CategoryUpsertDTO) (*model.CategoryDTO, error)
	GetByID(ctx context.Context, id int) (*model.CategoryDTO, error)
	List(ctx context.Context, query table.TableQuery) (table.TableListResult[model.CategoryDTO], error)
	Search(ctx context.Context, query dbutils.SearchQuery) (dbutils.SearchResult[model.CategoryDTO], error)
	Delete(ctx context.Context, id int) error
}

type categoryRepo struct {
	db             *generated.Client
	deps           *module.ModuleDeps[config.ModuleConfig]
	cfMgr          *customfields.Manager
	collectionRepo *repository.CollectionRepository
}

func NewCategoryRepository(db *generated.Client, deps *module.ModuleDeps[config.ModuleConfig], cfMgr *customfields.Manager) CategoryRepository {
	collectionRepo := repository.NewCollectionRepository(deps.DB)
	return &categoryRepo{
		db:             db,
		deps:           deps,
		cfMgr:          cfMgr,
		collectionRepo: collectionRepo,
	}
}

func (r *categoryRepo) upsertCollection(ctx context.Context, tx *generated.Tx, entity *generated.Category) (*generated.Category, error) {
	showIf := customfields.ShowIfCondition{
		Field: "categoryId",
		Op:    "eq",
		Value: entity.ID,
	}

	showIfJSON, err := json.Marshal(showIf)
	if err != nil {
		return nil, err
	}

	collectionSlug := "category-" + strconv.Itoa(entity.ID)
	collectionName := "Category"
	if entity.Name != nil && *entity.Name != "" {
		collectionName = *entity.Name
	}

	collectionGroup := "category"
	showIfValue := sql.NullString{String: string(showIfJSON), Valid: true}

	if entity.CollectionID != nil {
		integration := true
		_, err = r.collectionRepo.Update(ctx, *entity.CollectionID, &collectionSlug, &collectionName, &showIfValue, &integration, &collectionGroup)
		if err != nil {
			return nil, err
		}
		return entity, nil
	}

	collection, err := r.collectionRepo.Create(ctx, collectionSlug, collectionName, &showIfValue, true, &collectionGroup)
	if err != nil {
		return nil, err
	}

	return tx.Category.UpdateOneID(entity.ID).
		SetCollectionID(collection.ID).
		Save(ctx)
}

func (r *categoryRepo) Create(ctx context.Context, input *model.CategoryUpsertDTO) (*model.CategoryDTO, error) {
	tx, err := r.db.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			_ = tx.Commit()
		}
	}()

	dto := &input.DTO

	q := tx.Category.Create().
		SetNillableName(dto.Name)

	if input.Collections != nil && len(*input.Collections) > 0 {
		_, err = customfields.PrepareCustomFields(ctx,
			r.cfMgr,
			*input.Collections,
			dto.CustomFields,
			q,
			false,
		)
		if err != nil {
			return nil, err
		}
	}

	entity, err := q.Save(ctx)
	if err != nil {
		return nil, err
	}

	entity, err = r.upsertCollection(ctx, tx, entity)
	if err != nil {
		return nil, err
	}

	dto = mapper.MapAs[*generated.Category, *model.CategoryDTO](entity)

	err = relation.Upsert1(ctx, tx, "category", entity, &input.DTO, dto)
	if err != nil {
		return nil, err
	}

	_, err = relation.UpsertM2M(ctx, tx, "category", entity, input.DTO, dto)
	if err != nil {
		return nil, err
	}

	return dto, nil
}

func (r *categoryRepo) Update(ctx context.Context, input *model.CategoryUpsertDTO) (*model.CategoryDTO, error) {
	tx, err := r.db.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			_ = tx.Commit()
		}
	}()

	dto := &input.DTO

	q := tx.Category.UpdateOneID(dto.ID).
		SetNillableName(dto.Name)

	if input.Collections != nil && len(*input.Collections) > 0 {
		_, err = customfields.PrepareCustomFields(
			ctx,
			r.cfMgr,
			*input.Collections,
			dto.CustomFields,
			q,
			true,
		)
		if err != nil {
			return nil, err
		}
	}

	entity, err := q.Save(ctx)
	if err != nil {
		return nil, err
	}

	entity, err = r.upsertCollection(ctx, tx, entity)
	if err != nil {
		return nil, err
	}

	dto = mapper.MapAs[*generated.Category, *model.CategoryDTO](entity)

	err = relation.Upsert1(ctx, tx, "category", entity, &input.DTO, dto)
	if err != nil {
		return nil, err
	}

	_, err = relation.UpsertM2M(ctx, tx, "category", entity, input.DTO, dto)
	if err != nil {
		return nil, err
	}

	return dto, nil
}

func (r *categoryRepo) GetByID(ctx context.Context, id int) (*model.CategoryDTO, error) {
	q := r.db.Category.Query().
		Where(
			category.ID(id),
			category.DeletedAtIsNil(),
		)

	entity, err := q.Only(ctx)
	if err != nil {
		return nil, err
	}

	dto := mapper.MapAs[*generated.Category, *model.CategoryDTO](entity)
	return dto, nil
}

func (r *categoryRepo) List(ctx context.Context, query table.TableQuery) (table.TableListResult[model.CategoryDTO], error) {
	list, err := table.TableList(
		ctx,
		r.db.Category.Query().
			Where(category.DeletedAtIsNil()),
		query,
		category.Table,
		category.FieldID,
		category.FieldID,
		func(src []*generated.Category) []*model.CategoryDTO {
			return mapper.MapListAs[*generated.Category, *model.CategoryDTO](src)
		},
	)
	if err != nil {
		var zero table.TableListResult[model.CategoryDTO]
		return zero, err
	}
	return list, nil
}

func (r *categoryRepo) Search(ctx context.Context, query dbutils.SearchQuery) (dbutils.SearchResult[model.CategoryDTO], error) {
	return dbutils.Search(
		ctx,
		r.db.Category.Query().
			Where(category.DeletedAtIsNil()),
		[]string{
			dbutils.GetNormField(category.FieldName),
		},
		query,
		category.Table,
		category.FieldID,
		category.FieldID,
		category.Or,
		func(src []*generated.Category) []*model.CategoryDTO {
			return mapper.MapListAs[*generated.Category, *model.CategoryDTO](src)
		},
	)
}

func (r *categoryRepo) Delete(ctx context.Context, id int) error {
	return r.db.Category.UpdateOneID(id).
		SetDeletedAt(time.Now()).
		Exec(ctx)
}
