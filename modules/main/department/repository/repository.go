package repository

import (
	"context"

	"github.com/khiemnd777/andy_api/modules/main/config"
	"github.com/khiemnd777/andy_api/modules/main/department/model"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/department"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/departmentmember"
	"github.com/khiemnd777/andy_api/shared/mapper"
	"github.com/khiemnd777/andy_api/shared/module"
	"github.com/khiemnd777/andy_api/shared/utils/table"
)

type DepartmentRepository interface {
	Create(ctx context.Context, input model.DepartmentDTO) (*model.DepartmentDTO, error)
	Update(ctx context.Context, input model.DepartmentDTO) (*model.DepartmentDTO, error)
	GetByID(ctx context.Context, id int) (*model.DepartmentDTO, error)
	GetBySlug(ctx context.Context, slug string) (*model.DepartmentDTO, error)
	List(ctx context.Context, query table.TableQuery) (table.TableListResult[model.DepartmentDTO], error)
	ChildrenList(ctx context.Context, parentID int, query table.TableQuery) (table.TableListResult[model.DepartmentDTO], error)
	Delete(ctx context.Context, id int) error
	ExistsMembership(ctx context.Context, userID, deptID int) (bool, error)
	GetFirstDepartmentOfUser(ctx context.Context, userID int) (*model.DepartmentDTO, error)
}

type departmentRepo struct {
	db   *generated.Client
	deps *module.ModuleDeps[config.ModuleConfig]
}

func NewDepartmentRepository(db *generated.Client, deps *module.ModuleDeps[config.ModuleConfig]) DepartmentRepository {
	return &departmentRepo{db: db, deps: deps}
}

func (r *departmentRepo) Create(ctx context.Context, input model.DepartmentDTO) (*model.DepartmentDTO, error) {
	q := r.db.Department.Create().
		SetActive(input.Active).
		SetName(input.Name).
		SetNillableLogo(input.Logo).
		SetNillableAddress(input.Address).
		SetNillablePhoneNumber(input.PhoneNumber).
		SetNillableParentID(input.ParentID)

	entity, err := q.Save(ctx)

	if err != nil {
		return nil, err
	}

	departmentDTO := mapper.MapAs[*generated.Department, *model.DepartmentDTO](entity)

	return departmentDTO, nil
}

func (r *departmentRepo) Update(ctx context.Context, input model.DepartmentDTO) (*model.DepartmentDTO, error) {
	entity, err := r.db.Department.UpdateOneID(input.ID).
		SetActive(input.Active).
		SetName(input.Name).
		SetNillableLogo(input.Logo).
		SetNillableAddress(input.Address).
		SetNillablePhoneNumber(input.PhoneNumber).
		SetNillableParentID(input.ParentID).
		Save(ctx)

	if err != nil {
		return nil, err
	}

	departmentDTO := mapper.MapAs[*generated.Department, *model.DepartmentDTO](entity)

	return departmentDTO, nil
}

func (r *departmentRepo) GetByID(ctx context.Context, id int) (*model.DepartmentDTO, error) {
	entity, err := r.db.Department.Query().
		Where(department.ID(id), department.Deleted(false)).
		Only(ctx)

	if err != nil {
		return nil, err
	}

	departmentDTO := mapper.MapAs[*generated.Department, *model.DepartmentDTO](entity)

	return departmentDTO, nil
}

func (r *departmentRepo) GetBySlug(ctx context.Context, slug string) (*model.DepartmentDTO, error) {
	entity, err := r.db.Department.Query().
		Where(department.Slug(slug), department.Deleted(false)).
		Only(ctx)

	if err != nil {
		return nil, err
	}

	departmentDTO := mapper.MapAs[*generated.Department, *model.DepartmentDTO](entity)

	return departmentDTO, nil
}

func (r *departmentRepo) List(ctx context.Context, query table.TableQuery) (table.TableListResult[model.DepartmentDTO], error) {
	list, err := table.TableListV2(
		ctx,
		r.db.Department.Query().
			Where(department.Deleted(false)),
		query,
		department.Table,
		department.FieldID,
		department.FieldID,
		func(q *generated.DepartmentQuery) *generated.DepartmentQuery {
			return q
		},
		func(src []*generated.Department) []*model.DepartmentDTO {
			return mapper.MapListAs[*generated.Department, *model.DepartmentDTO](src)
		},
	)
	if err != nil {
		var zero table.TableListResult[model.DepartmentDTO]
		return zero, err
	}
	return list, nil
}

func (r *departmentRepo) ChildrenList(ctx context.Context, parentID int, query table.TableQuery) (table.TableListResult[model.DepartmentDTO], error) {
	list, err := table.TableListV2(
		ctx,
		r.db.Department.Query().
			Where(
				department.Deleted(false),
				department.ParentIDEQ(parentID),
			),
		query,
		department.Table,
		department.FieldID,
		department.FieldID,
		func(q *generated.DepartmentQuery) *generated.DepartmentQuery {
			return q
		},
		func(src []*generated.Department) []*model.DepartmentDTO {
			return mapper.MapListAs[*generated.Department, *model.DepartmentDTO](src)
		},
	)
	if err != nil {
		var zero table.TableListResult[model.DepartmentDTO]
		return zero, err
	}
	return list, nil
}

func (r *departmentRepo) Delete(ctx context.Context, id int) error {
	return r.db.Department.UpdateOneID(id).
		SetDeleted(true).
		Exec(ctx)
}

func (r *departmentRepo) ExistsMembership(ctx context.Context, userID, deptID int) (bool, error) {
	n, err := r.db.DepartmentMember.
		Query().
		Where(
			departmentmember.UserID(userID),
			departmentmember.DepartmentID(deptID),
		).
		Count(ctx)
	return n > 0, err
}

func (r *departmentRepo) GetFirstDepartmentOfUser(ctx context.Context, userID int) (*model.DepartmentDTO, error) {
	dm, err := r.db.DepartmentMember.Query().
		Where(departmentmember.UserID(userID)).
		Order(departmentmember.ByCreatedAt()).
		First(ctx)

	if err != nil {
		return nil, err
	}

	d, err := dm.QueryDepartment().
		Where(department.Deleted(false)).
		Only(ctx)

	if err != nil {
		return nil, err
	}

	res := mapper.MapAs[*generated.Department, *model.DepartmentDTO](d)

	return res, nil
}
