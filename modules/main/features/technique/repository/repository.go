package repository

import (
	"context"
	"time"

	"github.com/khiemnd777/andy_api/modules/main/config"
	model "github.com/khiemnd777/andy_api/modules/main/features/__model"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/technique"
	dbutils "github.com/khiemnd777/andy_api/shared/db/utils"
	"github.com/khiemnd777/andy_api/shared/mapper"
	"github.com/khiemnd777/andy_api/shared/module"
	"github.com/khiemnd777/andy_api/shared/utils/table"
)

type TechniqueRepository interface {
	Create(ctx context.Context, input model.TechniqueDTO) (*model.TechniqueDTO, error)
	Update(ctx context.Context, input model.TechniqueDTO) (*model.TechniqueDTO, error)
	GetByID(ctx context.Context, id int) (*model.TechniqueDTO, error)
	List(ctx context.Context, categoryID *int, query table.TableQuery) (table.TableListResult[model.TechniqueDTO], error)
	Search(ctx context.Context, categoryID *int, query dbutils.SearchQuery) (dbutils.SearchResult[model.TechniqueDTO], error)
	Delete(ctx context.Context, id int) error
}

type techniqueRepo struct {
	db   *generated.Client
	deps *module.ModuleDeps[config.ModuleConfig]
}

func NewTechniqueRepository(db *generated.Client, deps *module.ModuleDeps[config.ModuleConfig]) TechniqueRepository {
	return &techniqueRepo{db: db, deps: deps}
}

func (r *techniqueRepo) Create(ctx context.Context, input model.TechniqueDTO) (*model.TechniqueDTO, error) {
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

	entity, err := tx.Technique.Create().
		SetNillableCategoryID(input.CategoryID).
		SetNillableName(input.Name).
		Save(ctx)
	if err != nil {
		return nil, err
	}

	dto := mapper.MapAs[*generated.Technique, *model.TechniqueDTO](entity)
	return dto, nil
}

func (r *techniqueRepo) Update(ctx context.Context, input model.TechniqueDTO) (*model.TechniqueDTO, error) {
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

	entity, err := tx.Technique.UpdateOneID(input.ID).
		SetNillableCategoryID(input.CategoryID).
		SetNillableName(input.Name).
		Save(ctx)
	if err != nil {
		return nil, err
	}

	dto := mapper.MapAs[*generated.Technique, *model.TechniqueDTO](entity)
	return dto, nil
}

func (r *techniqueRepo) GetByID(ctx context.Context, id int) (*model.TechniqueDTO, error) {
	entity, err := r.db.Technique.Query().
		Where(
			technique.ID(id),
			technique.DeletedAtIsNil(),
		).
		Only(ctx)
	if err != nil {
		return nil, err
	}

	dto := mapper.MapAs[*generated.Technique, *model.TechniqueDTO](entity)
	return dto, nil
}

func (r *techniqueRepo) List(ctx context.Context, categoryID *int, query table.TableQuery) (table.TableListResult[model.TechniqueDTO], error) {
	q := r.db.Technique.Query().
		Where(technique.DeletedAtIsNil())
	if categoryID != nil {
		q = q.Where(technique.CategoryIDEQ(*categoryID))
	}

	list, err := table.TableList(
		ctx,
		q,
		query,
		technique.Table,
		technique.FieldID,
		technique.FieldID,
		func(src []*generated.Technique) []*model.TechniqueDTO {
			return mapper.MapListAs[*generated.Technique, *model.TechniqueDTO](src)
		},
	)
	if err != nil {
		var zero table.TableListResult[model.TechniqueDTO]
		return zero, err
	}
	return list, nil
}

func (r *techniqueRepo) Search(ctx context.Context, categoryID *int, query dbutils.SearchQuery) (dbutils.SearchResult[model.TechniqueDTO], error) {
	q := r.db.Technique.Query().
		Where(technique.DeletedAtIsNil())
	if categoryID != nil {
		q = q.Where(technique.CategoryIDEQ(*categoryID))
	}

	return dbutils.Search(
		ctx,
		q,
		[]string{
			dbutils.GetNormField(technique.FieldName),
		},
		query,
		technique.Table,
		technique.FieldID,
		technique.FieldID,
		technique.Or,
		func(src []*generated.Technique) []*model.TechniqueDTO {
			return mapper.MapListAs[*generated.Technique, *model.TechniqueDTO](src)
		},
	)
}

func (r *techniqueRepo) Delete(ctx context.Context, id int) error {
	return r.db.Technique.UpdateOneID(id).
		SetDeletedAt(time.Now()).
		Exec(ctx)
}
