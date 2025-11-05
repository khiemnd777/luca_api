package service

import (
	"context"
	"fmt"

	"github.com/khiemnd777/andy_api/modules/main/config"
	"github.com/khiemnd777/andy_api/modules/main/department/model"
	"github.com/khiemnd777/andy_api/modules/main/department/repository"
	"github.com/khiemnd777/andy_api/shared/cache"
	"github.com/khiemnd777/andy_api/shared/logger"
	"github.com/khiemnd777/andy_api/shared/mapper"
	"github.com/khiemnd777/andy_api/shared/module"
)

type DepartmentService interface {
	Create(ctx context.Context, input model.DepartmentDTO) (*model.DepartmentDTO, error)
	Update(ctx context.Context, input model.DepartmentDTO, userID int) (*model.DepartmentDTO, error)
	GetByID(ctx context.Context, id int) (*model.DepartmentDTO, error)
	GetBySlug(ctx context.Context, slug string) (*model.DepartmentDTO, error)
	List(ctx context.Context, limit, offset int) ([]*model.DepartmentDTO, int, error)
	ChildrenList(ctx context.Context, parentID, limit, offset int) ([]*model.DepartmentDTO, int, error)
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

func keyDeptList(limit, offset int) string {
	return fmt.Sprintf("department:list:l%d:o%d", limit, offset)
}

func keyDeptChildren(parentID, limit, offset int) string {
	return fmt.Sprintf("department:children:p%d:l%d:o%d", parentID, limit, offset)
}

func keyMyFirstDept(userID int) string {
	return fmt.Sprintf("department:first_of_user:%d", userID)
}

func invalidateDept(id int, slug string) {
	cache.InvalidateKeys(
		keyDept(id),
		keyDeptSlug(slug),
		"department:list:*",
		"department:children:*",
	)
}

func (s *departmentService) Create(ctx context.Context, input model.DepartmentDTO) (*model.DepartmentDTO, error) {
	res, err := s.repo.Create(ctx, input)
	if err == nil {
		invalidateDept(res.ID, *res.Slug)
	}
	return res, err
}

func (s *departmentService) Update(ctx context.Context, input model.DepartmentDTO, userID int) (*model.DepartmentDTO, error) {
	logger.Debug("HERE 1")
	res, err := s.repo.Update(ctx, input)
	if err == nil {
		invalidateDept(res.ID, *res.Slug)
		cache.InvalidateKeys(keyMyFirstDept(userID))
	}
	logger.Debug("HERE 2")
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

func (s *departmentService) List(ctx context.Context, limit, offset int) ([]*model.DepartmentDTO, int, error) {
	var totalRes = 0
	res, err := cache.GetList(keyDeptList(limit, offset), cache.TTLMedium, func() ([]*model.DepartmentDTO, error) {
		res, total, err := s.repo.List(ctx, limit, offset)
		totalRes = total
		return res, err
	})
	if err != nil {
		return nil, 0, err
	}
	return res, totalRes, nil
}

func (s *departmentService) ChildrenList(ctx context.Context, parentID, limit, offset int) ([]*model.DepartmentDTO, int, error) {
	var totalRes = 0
	res, err := cache.GetList(keyDeptChildren(parentID, limit, offset), cache.TTLMedium, func() ([]*model.DepartmentDTO, error) {
		res, total, err := s.repo.ChildrenList(ctx, parentID, limit, offset)
		totalRes = total
		return res, err
	})
	if err != nil {
		return nil, 0, err
	}
	return res, totalRes, nil
}

func (s *departmentService) Delete(ctx context.Context, id int) error {
	res, err := s.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	invalidateDept(id, *res.Slug)
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
