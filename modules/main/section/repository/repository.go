package repository

import (
	"context"
	"time"

	"github.com/khiemnd777/andy_api/modules/main/config"
	"github.com/khiemnd777/andy_api/modules/main/section/model"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/section"
	dbutils "github.com/khiemnd777/andy_api/shared/db/utils"
	"github.com/khiemnd777/andy_api/shared/mapper"
	"github.com/khiemnd777/andy_api/shared/module"
	"github.com/khiemnd777/andy_api/shared/utils/table"
)

type SectionRepository interface {
	Create(ctx context.Context, input model.SectionDTO) (*model.SectionDTO, error)
	Update(ctx context.Context, input model.SectionDTO) (*model.SectionDTO, error)
	GetByID(ctx context.Context, id int) (*model.SectionDTO, error)
	List(ctx context.Context, query table.TableQuery) (table.TableListResult[model.SectionDTO], error)
	Search(ctx context.Context, query dbutils.SearchQuery) (dbutils.SearchResult[model.SectionDTO], error)
	Delete(ctx context.Context, id int) error
}

type sectionRepo struct {
	db   *generated.Client
	deps *module.ModuleDeps[config.ModuleConfig]
}

func NewSectionRepository(db *generated.Client, deps *module.ModuleDeps[config.ModuleConfig]) SectionRepository {
	return &sectionRepo{db: db, deps: deps}
}

func (r *sectionRepo) Create(ctx context.Context, input model.SectionDTO) (*model.SectionDTO, error) {
	q := r.db.Section.Create().
		SetActive(input.Active).
		SetName(input.Name).
		SetNillableCode(input.Code).
		SetDescription(input.Description)

	entity, err := q.Save(ctx)
	if err != nil {
		return nil, err
	}

	dto := mapper.MapAs[*generated.Section, *model.SectionDTO](entity)
	return dto, nil
}

func (r *sectionRepo) Update(ctx context.Context, input model.SectionDTO) (*model.SectionDTO, error) {
	entity, err := r.db.Section.UpdateOneID(input.ID).
		SetActive(input.Active).
		SetName(input.Name).
		SetNillableCode(input.Code).
		SetDescription(input.Description).
		Save(ctx)
	if err != nil {
		return nil, err
	}

	dto := mapper.MapAs[*generated.Section, *model.SectionDTO](entity)
	return dto, nil
}

func (r *sectionRepo) GetByID(ctx context.Context, id int) (*model.SectionDTO, error) {
	entity, err := r.db.Section.Query().
		Where(
			section.ID(id),
			section.DeletedAtIsNil(),
		).
		Only(ctx)
	if err != nil {
		return nil, err
	}

	dto := mapper.MapAs[*generated.Section, *model.SectionDTO](entity)
	return dto, nil
}

func (r *sectionRepo) List(ctx context.Context, query table.TableQuery) (table.TableListResult[model.SectionDTO], error) {
	list, err := table.TableList(
		ctx,
		r.db.Section.Query().
			Where(section.DeletedAtIsNil()),
		query,
		section.Table,
		section.FieldID,
		section.FieldID,
		func(src []*generated.Section) []*model.SectionDTO {
			mapped := mapper.MapListAs[*generated.Section, *model.SectionDTO](src)
			return mapped
		},
	)
	if err != nil {
		var zero table.TableListResult[model.SectionDTO]
		return zero, err
	}
	return list, nil
}

func (r *sectionRepo) Search(ctx context.Context, query dbutils.SearchQuery) (dbutils.SearchResult[model.SectionDTO], error) {
	return dbutils.Search(
		ctx,
		r.db.Section.Query().
			Where(section.DeletedAtIsNil()),
		[]string{
			dbutils.GetNormField(section.FieldName),
		},
		query,
		section.Table,
		section.FieldID,
		section.FieldID,
		section.Or,
		func(src []*generated.Section) []*model.SectionDTO {
			mapped := mapper.MapListAs[*generated.Section, *model.SectionDTO](src)
			return mapped
		},
	)
}

func (r *sectionRepo) Delete(ctx context.Context, id int) error {
	return r.db.Section.UpdateOneID(id).
		SetDeletedAt(time.Now()).
		Exec(ctx)
}
