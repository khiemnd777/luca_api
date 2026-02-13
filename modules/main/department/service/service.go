package service

import (
	"context"
	"fmt"

	"github.com/khiemnd777/andy_api/modules/main/config"
	"github.com/khiemnd777/andy_api/modules/main/department/model"
	"github.com/khiemnd777/andy_api/modules/main/department/repository"
	"github.com/khiemnd777/andy_api/shared/cache"
	"github.com/khiemnd777/andy_api/shared/mapper"
	"github.com/khiemnd777/andy_api/shared/module"
	"github.com/khiemnd777/andy_api/shared/utils/table"
)

type DepartmentService interface {
	Create(ctx context.Context, input model.DepartmentDTO) (*model.DepartmentDTO, error)
	Update(ctx context.Context, input model.DepartmentDTO, userID int) (*model.DepartmentDTO, error)
	GetByID(ctx context.Context, id int) (*model.DepartmentDTO, error)
	GetBySlug(ctx context.Context, slug string) (*model.DepartmentDTO, error)
	List(ctx context.Context, query table.TableQuery) (table.TableListResult[model.DepartmentDTO], error)
	ChildrenList(ctx context.Context, parentID int, query table.TableQuery) (table.TableListResult[model.DepartmentDTO], error)
	Delete(ctx context.Context, id int) error
	GetFirstDepartmentOfUser(ctx context.Context, userID int) (*model.DepartmentDTO, error)
}

type departmentService struct {
	repo repository.DepartmentRepository
	deps *module.ModuleDeps[config.ModuleConfig]
}

func NewDepartmentService(repo repository.DepartmentRepository, deps *module.ModuleDeps[config.ModuleConfig]) DepartmentService {
	return &departmentService{repo: repo, deps: deps}
}

func keyDept(id int) string {
	return fmt.Sprintf("department:%d", id)
}

func keyDeptSlug(slug string) string {
	return fmt.Sprintf("department:slug:%s", slug)
}

func keyDeptList(query table.TableQuery) string {
	orderBy := ""
	if query.OrderBy != nil {
		orderBy = *query.OrderBy
	}
	return fmt.Sprintf(
		"department:list:l%d:p%d:o%d:ob%s:d%s",
		query.Limit,
		query.Page,
		query.Offset,
		orderBy,
		query.Direction,
	)
}

func keyDeptChildren(parentID int, query table.TableQuery) string {
	orderBy := ""
	if query.OrderBy != nil {
		orderBy = *query.OrderBy
	}
	return fmt.Sprintf(
		"department:children:p%d:l%d:p%d:o%d:ob%s:d%s",
		parentID,
		query.Limit,
		query.Page,
		query.Offset,
		orderBy,
		query.Direction,
	)
}

func keyMyFirstDept(userID int) string {
	return fmt.Sprintf("department:first_of_user:%d", userID)
}

func invalidateDept(id int) {
	cache.InvalidateKeys(
		keyDept(id),
		"department:list:*",
		"department:children:*",
	)
}

func (s *departmentService) Create(ctx context.Context, input model.DepartmentDTO) (*model.DepartmentDTO, error) {
	res, err := s.repo.Create(ctx, input)
	if err == nil {
		invalidateDept(res.ID)
	}
	return res, err
}

func (s *departmentService) Update(ctx context.Context, input model.DepartmentDTO, userID int) (*model.DepartmentDTO, error) {
	res, err := s.repo.Update(ctx, input)
	if err == nil {
		invalidateDept(res.ID)
		cache.InvalidateKeys(keyMyFirstDept(userID))
	}
	return res, err
}

func (s *departmentService) GetByID(ctx context.Context, id int) (*model.DepartmentDTO, error) {
	return cache.Get(keyDept(id), cache.TTLLong, func() (*model.DepartmentDTO, error) {
		return s.repo.GetByID(ctx, id)
	})
}

func (s *departmentService) GetBySlug(ctx context.Context, slug string) (*model.DepartmentDTO, error) {
	return cache.Get(keyDeptSlug(slug), cache.TTLLong, func() (*model.DepartmentDTO, error) {
		return s.repo.GetBySlug(ctx, slug)
	})
}

func (s *departmentService) List(ctx context.Context, query table.TableQuery) (table.TableListResult[model.DepartmentDTO], error) {
	type boxed = table.TableListResult[model.DepartmentDTO]
	key := keyDeptList(query)

	ptr, err := cache.Get(key, cache.TTLMedium, func() (*boxed, error) {
		res, e := s.repo.List(ctx, query)
		if e != nil {
			return nil, e
		}
		return &res, nil
	})
	if err != nil {
		var zero boxed
		return zero, err
	}
	return *ptr, nil
}

func (s *departmentService) ChildrenList(ctx context.Context, parentID int, query table.TableQuery) (table.TableListResult[model.DepartmentDTO], error) {
	type boxed = table.TableListResult[model.DepartmentDTO]
	key := keyDeptChildren(parentID, query)

	ptr, err := cache.Get(key, cache.TTLMedium, func() (*boxed, error) {
		res, e := s.repo.ChildrenList(ctx, parentID, query)
		if e != nil {
			return nil, e
		}
		return &res, nil
	})
	if err != nil {
		var zero boxed
		return zero, err
	}
	return *ptr, nil
}

func (s *departmentService) Delete(ctx context.Context, id int) error {
	_, err := s.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	invalidateDept(id)
	return nil
}

func (s *departmentService) GetFirstDepartmentOfUser(ctx context.Context, userID int) (*model.DepartmentDTO, error) {
	key := keyMyFirstDept(userID)

	res, err := cache.Get(key, cache.TTLMedium, func() (*model.DepartmentDTO, error) {
		e, err := s.repo.GetFirstDepartmentOfUser(ctx, userID)
		if err != nil {
			return nil, err
		}
		return mapper.Map(&e), nil
	})
	if err != nil {
		return nil, err
	}
	return res, nil
}
