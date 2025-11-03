// scripts/create_module/templates/service_service.go.tmpl
package service

import (
	"context"
	"fmt"

	"github.com/khiemnd777/andy_api/modules/folder/config"
	"github.com/khiemnd777/andy_api/modules/folder/repository"
	"github.com/khiemnd777/andy_api/shared/cache"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/module"
)

type FolderService struct {
	repo *repository.FolderRepository
	deps *module.ModuleDeps[config.ModuleConfig]
}

func NewFolderService(repo *repository.FolderRepository, deps *module.ModuleDeps[config.ModuleConfig]) *FolderService {
	return &FolderService{
		repo: repo,
		deps: deps,
	}
}

func (s *FolderService) Create(ctx context.Context, userId int, input *generated.Folder) (*generated.Folder, error) {
	key := fmt.Sprintf("user:%d:folder:list", userId)
	keyListFirstPage := fmt.Sprintf("user:%d:folder:list:first-page", userId)

	var folderResult *generated.Folder

	err := cache.UpdateManyAndInvalidate([]string{key, keyListFirstPage}, func() error {
		folder, error := s.repo.Create(ctx, input)
		folderResult = folder
		return error
	})

	return folderResult, err
}

func (s *FolderService) Get(ctx context.Context, id, userId int) (*generated.Folder, error) {
	key := fmt.Sprintf("user:%d:folder:id:%d", userId, id)
	return cache.Get(key, cache.TTLShort, func() (*generated.Folder, error) {
		return s.repo.FindByID(ctx, id)
	})
}

func (s *FolderService) List(ctx context.Context, userId int) ([]*generated.Folder, error) {
	key := fmt.Sprintf("user:%d:folder:list", userId)
	return cache.GetList(key, cache.TTLLong, func() ([]*generated.Folder, error) {
		return s.repo.ListByUser(ctx, userId)
	})
}

func (s *FolderService) ListPaginated(ctx context.Context, userId, page, limit int) ([]*generated.Folder, bool, error) {
	offset := (page - 1) * limit
	if page == 1 {
		keyListFirstPage := fmt.Sprintf("user:%d:folder:list:first-page", userId)
		return cache.GetListWithHasMore(keyListFirstPage, cache.TTLLong, func() ([]*generated.Folder, bool, error) {
			return s.repo.ListByUserPaginated(ctx, userId, limit, offset)
		})
	}
	return s.repo.ListByUserPaginated(ctx, userId, limit, offset)
}

func (s *FolderService) Update(ctx context.Context, id, userId int, input *generated.Folder) (*generated.Folder, error) {
	keyFolderSingle := fmt.Sprintf("user:%d:folder:id:%d", userId, id)
	keyFolderList := fmt.Sprintf("user:%d:folder:list", userId)
	keyListFirstPage := fmt.Sprintf("user:%d:folder:list:first-page", userId)

	var folderResult *generated.Folder

	err := cache.UpdateManyAndInvalidate([]string{keyFolderList, keyFolderSingle, keyListFirstPage}, func() error {
		folder, err := s.repo.Update(ctx, id, input)
		folderResult = folder
		return err
	})
	return folderResult, err
}

func (s *FolderService) Delete(ctx context.Context, id, userId int) error {
	keyFolderSingle := fmt.Sprintf("user:%d:folder:id:%d", userId, id)
	keyFolderList := fmt.Sprintf("user:%d:folder:list", userId)
	keyListFirstPage := fmt.Sprintf("user:%d:folder:list:first-page", userId)
	return cache.UpdateManyAndInvalidate([]string{keyFolderSingle, keyFolderList, keyListFirstPage}, func() error {
		return s.repo.SoftDelete(ctx, id)
	})
}
