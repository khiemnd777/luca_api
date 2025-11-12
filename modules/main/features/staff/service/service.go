package service

import (
	"context"
	"fmt"

	"github.com/khiemnd777/andy_api/modules/main/config"
	model "github.com/khiemnd777/andy_api/modules/main/features/__model"
	"github.com/khiemnd777/andy_api/modules/main/features/staff/repository"
	"github.com/khiemnd777/andy_api/shared/cache"
	dbutils "github.com/khiemnd777/andy_api/shared/db/utils"
	"github.com/khiemnd777/andy_api/shared/module"
	"github.com/khiemnd777/andy_api/shared/utils/table"
)

type StaffService interface {
	Create(ctx context.Context, input model.StaffDTO) (*model.StaffDTO, error)
	Update(ctx context.Context, input model.StaffDTO) (*model.StaffDTO, error)
	ChangePassword(ctx context.Context, id int, newPassword string) error
	GetByID(ctx context.Context, id int) (*model.StaffDTO, error)
	CheckPhoneExists(ctx context.Context, userID int, phone string) (bool, error)
	CheckEmailExists(ctx context.Context, userID int, email string) (bool, error)
	List(ctx context.Context, query table.TableQuery) (table.TableListResult[model.StaffDTO], error)
	ListBySectionID(ctx context.Context, sectionID int, query table.TableQuery) (table.TableListResult[model.StaffDTO], error)
	Search(ctx context.Context, query dbutils.SearchQuery) (dbutils.SearchResult[model.StaffDTO], error)
	Delete(ctx context.Context, id int) error
}

type staffService struct {
	repo repository.StaffRepository
	deps *module.ModuleDeps[config.ModuleConfig]
}

func NewStaffService(repo repository.StaffRepository, deps *module.ModuleDeps[config.ModuleConfig]) StaffService {
	return &staffService{repo: repo, deps: deps}
}

func kStaffByID(id int) string {
	return fmt.Sprintf("staff:id:%d", id)
}

func kStaffAll() []string {
	return []string{
		kStaffListAll(),
		kStaffSearchAll(),
		kStaffSectionAll(),
	}
}

func kStaffListAll() string {
	return "staff:list:*"
}

func kSectionStaffAll(staffID int) string {
	return fmt.Sprintf("section:staff:%d:*", staffID)
}

func kStaffSearchAll() string {
	return "staff:search:*"
}

func kStaffSectionAll() string {
	return "staff:section:*"
}

func kStaffSectionList(staffID int) string {
	return fmt.Sprintf("section:staff:%d:*", staffID)
}

func kUserRoleList(staffID int) string {
	return fmt.Sprintf("rbac:roles:user:%d:*", staffID)
}

func kStaffList(q table.TableQuery) string {
	orderBy := ""
	if q.OrderBy != nil {
		orderBy = *q.OrderBy
	}
	return fmt.Sprintf("staff:list:l%d:p%d:o%s:d%s", q.Limit, q.Page, orderBy, q.Direction)
}

func kSectionStaffList(sectionID int, q table.TableQuery) string {
	orderBy := ""
	if q.OrderBy != nil {
		orderBy = *q.OrderBy
	}
	return fmt.Sprintf("staff:section:%d:list:l%d:p%d:o%s:d%s", sectionID, q.Limit, q.Page, orderBy, q.Direction)
}

func kStaffSearch(q dbutils.SearchQuery) string {
	orderBy := ""
	if q.OrderBy != nil {
		orderBy = *q.OrderBy
	}
	return fmt.Sprintf("staff:search:k%s:l%d:p%d:o%s:d%s", q.Keyword, q.Limit, q.Page, orderBy, q.Direction)
}

func (s *staffService) Create(ctx context.Context, input model.StaffDTO) (*model.StaffDTO, error) {
	dto, err := s.repo.Create(ctx, input)
	if err != nil {
		return nil, err
	}

	cache.InvalidateKeys(kStaffAll()...)
	if dto != nil && dto.ID > 0 {
		cache.InvalidateKeys(kStaffByID(dto.ID), kStaffSectionList(dto.ID), kUserRoleList(dto.ID), kSectionStaffAll(dto.ID))
	}
	return dto, nil
}

func (s *staffService) Update(ctx context.Context, input model.StaffDTO) (*model.StaffDTO, error) {
	dto, err := s.repo.Update(ctx, input)
	if err != nil {
		return nil, err
	}

	if dto != nil {
		cache.InvalidateKeys(kStaffByID(dto.ID), kStaffSectionList(dto.ID), kUserRoleList(dto.ID), kSectionStaffAll(dto.ID))
	}
	cache.InvalidateKeys(kStaffAll()...)
	return dto, nil
}

func (s *staffService) ChangePassword(ctx context.Context, id int, newPassword string) error {
	return s.repo.ChangePassword(ctx, id, newPassword)
}

func (s *staffService) GetByID(ctx context.Context, id int) (*model.StaffDTO, error) {
	return cache.Get(kStaffByID(id), cache.TTLMedium, func() (*model.StaffDTO, error) {
		return s.repo.GetByID(ctx, id)
	})
}

func (s *staffService) List(ctx context.Context, q table.TableQuery) (table.TableListResult[model.StaffDTO], error) {
	type boxed = table.TableListResult[model.StaffDTO]
	key := kStaffList(q)

	ptr, err := cache.Get(key, cache.TTLMedium, func() (*boxed, error) {
		res, e := s.repo.List(ctx, q)
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

func (s *staffService) ListBySectionID(ctx context.Context, sectionID int, query table.TableQuery) (table.TableListResult[model.StaffDTO], error) {
	type boxed = table.TableListResult[model.StaffDTO]
	key := kSectionStaffList(sectionID, query)

	ptr, err := cache.Get(key, cache.TTLMedium, func() (*boxed, error) {
		res, e := s.repo.ListBySectionID(ctx, sectionID, query)
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

func (s *staffService) CheckPhoneExists(ctx context.Context, userID int, phone string) (bool, error) {
	return s.repo.CheckPhoneExists(ctx, userID, phone)
}

func (s *staffService) CheckEmailExists(ctx context.Context, userID int, email string) (bool, error) {
	return s.repo.CheckEmailExists(ctx, userID, email)
}

func (s *staffService) Delete(ctx context.Context, id int) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	cache.InvalidateKeys(kStaffAll()...)
	cache.InvalidateKeys(kStaffByID(id), kStaffSectionList(id), kUserRoleList(id), kSectionStaffAll(id))
	return nil
}

func (s *staffService) Search(ctx context.Context, q dbutils.SearchQuery) (dbutils.SearchResult[model.StaffDTO], error) {
	type boxed = dbutils.SearchResult[model.StaffDTO]
	key := kStaffSearch(q)

	ptr, err := cache.Get(key, cache.TTLMedium, func() (*boxed, error) {
		res, e := s.repo.Search(ctx, q)
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
