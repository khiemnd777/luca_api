package repository

import (
	"context"
	"time"

	"github.com/khiemnd777/andy_api/modules/main/config"
	model "github.com/khiemnd777/andy_api/modules/main/features/__model"
	relation "github.com/khiemnd777/andy_api/modules/main/features/__relation/policy"
	processrepo "github.com/khiemnd777/andy_api/modules/main/features/process/repository"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/section"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/staff"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/staffsection"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/user"
	dbutils "github.com/khiemnd777/andy_api/shared/db/utils"
	"github.com/khiemnd777/andy_api/shared/mapper"
	"github.com/khiemnd777/andy_api/shared/metadata/customfields"
	"github.com/khiemnd777/andy_api/shared/module"
	"github.com/khiemnd777/andy_api/shared/utils/table"
)

type SectionRepository interface {
	Create(ctx context.Context, input model.SectionDTO) (*model.SectionDTO, error)
	Update(ctx context.Context, input model.SectionDTO) (*model.SectionDTO, error)
	GetByID(ctx context.Context, id int) (*model.SectionDTO, error)
	List(ctx context.Context, deptID int, query table.TableQuery) (table.TableListResult[model.SectionDTO], error)
	ListByStaffID(ctx context.Context, staffID int, query table.TableQuery) (table.TableListResult[model.SectionDTO], error)
	Search(ctx context.Context, deptID int, query dbutils.SearchQuery) (dbutils.SearchResult[model.SectionDTO], error)
	Delete(ctx context.Context, id int) error
}

type sectionRepo struct {
	db          *generated.Client
	deps        *module.ModuleDeps[config.ModuleConfig]
	processRepo processrepo.ProcessRepository
	cfMgr       *customfields.Manager
}

func NewSectionRepository(db *generated.Client, deps *module.ModuleDeps[config.ModuleConfig], cfMgr *customfields.Manager) SectionRepository {
	return &sectionRepo{
		db:          db,
		deps:        deps,
		processRepo: processrepo.NewProcessRepository(db, deps, cfMgr),
		cfMgr:       cfMgr,
	}
}

func (r *sectionRepo) Create(ctx context.Context, input model.SectionDTO) (*model.SectionDTO, error) {
	return dbutils.WithTx(ctx, r.db, func(tx *generated.Tx) (*model.SectionDTO, error) {
		q := r.db.Section.Create().
			SetDepartmentID(input.DepartmentID).
			SetNillableLeaderID(input.LeaderID).
			SetNillableLeaderName(input.LeaderName).
			SetActive(input.Active).
			SetName(input.Name).
			SetNillableCode(input.Code).
			SetNillableColor(input.Color).
			SetDescription(input.Description)

		// custom fields
		_, err := customfields.PrepareCustomFields(ctx,
			r.cfMgr,
			[]string{"section"},
			input.CustomFields,
			q,
			false,
		)
		if err != nil {
			return nil, err
		}

		entity, err := q.Save(ctx)
		if err != nil {
			return nil, err
		}

		dto := mapper.MapAs[*generated.Section, *model.SectionDTO](entity)

		_, err = relation.UpsertM2M(ctx, tx, "sections_processes", entity, input, dto)
		if err != nil {
			return nil, err
		}

		return dto, nil
	})
}

func (r *sectionRepo) Update(ctx context.Context, input model.SectionDTO) (*model.SectionDTO, error) {
	return dbutils.WithTx(ctx, r.db, func(tx *generated.Tx) (*model.SectionDTO, error) {
		q := r.db.Section.UpdateOneID(input.ID).
			SetDepartmentID(input.DepartmentID).
			SetNillableLeaderID(input.LeaderID).
			SetNillableLeaderName(input.LeaderName).
			SetActive(input.Active).
			SetName(input.Name).
			SetNillableCode(input.Code).
			SetNillableColor(input.Color).
			SetDescription(input.Description)

		// custom fields
		_, err := customfields.PrepareCustomFields(ctx,
			r.cfMgr,
			[]string{"section"},
			input.CustomFields,
			q,
			false,
		)
		if err != nil {
			return nil, err
		}

		entity, err := q.Save(ctx)
		if err != nil {
			return nil, err
		}

		dto := mapper.MapAs[*generated.Section, *model.SectionDTO](entity)

		_, err = relation.UpsertM2M(ctx, tx, "sections_processes", entity, input, dto)
		if err != nil {
			return nil, err
		}

		return dto, nil
	})
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

func (r *sectionRepo) List(ctx context.Context, deptID int, query table.TableQuery) (table.TableListResult[model.SectionDTO], error) {
	list, err := table.TableList(
		ctx,
		r.db.Section.Query().
			Where(
				section.DeletedAtIsNil(),
				section.DepartmentIDEQ(deptID),
			),
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

func (r *sectionRepo) ListByStaffID(ctx context.Context, staffID int, query table.TableQuery) (table.TableListResult[model.SectionDTO], error) {
	list, err := table.TableList(
		ctx,
		r.db.Section.Query().
			Where(
				section.DeletedAtIsNil(),
				section.HasStaffsWith(staffsection.HasStaffWith(staff.HasUserWith(user.IDEQ(staffID)))),
			),
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

func (r *sectionRepo) Search(ctx context.Context, deptID int, query dbutils.SearchQuery) (dbutils.SearchResult[model.SectionDTO], error) {
	return dbutils.Search(
		ctx,
		r.db.Section.Query().
			Where(
				section.DeletedAtIsNil(),
				section.DepartmentIDEQ(deptID),
			),
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
