// scripts/create_module/templates/service_service.go.tmpl
package service

import (
	"context"
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"

	"github.com/khiemnd777/andy_api/modules/photo/config"
	"github.com/khiemnd777/andy_api/modules/photo/repository"
	batchUtil "github.com/khiemnd777/andy_api/shared/batch"
	"github.com/khiemnd777/andy_api/shared/cache"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/module"
	"github.com/khiemnd777/andy_api/shared/utils"
)

type PhotoService struct {
	repo *repository.PhotoRepository
	deps *module.ModuleDeps[config.ModuleConfig]
}

func NewPhotoService(repo *repository.PhotoRepository, deps *module.ModuleDeps[config.ModuleConfig]) *PhotoService {
	return &PhotoService{
		repo: repo,
		deps: deps,
	}
}

func (s *PhotoService) UploadAndSave(ctx context.Context, fileHeader *multipart.FileHeader, userId int, folderID *int, meta map[string]any) (*generated.Photo, error) {
	/*
		storage/photo/
		├── original/
		│   └── abc.jpg
		├── medium/
		│   └── abc.jpg
		└── thumbnail/
			└── abc.jpg
	*/

	ok, ext := IsSupportedImageType(fileHeader)

	if !ok {
		return nil, fmt.Errorf("unsupported image format: %s", ext)
	}

	photoPath := s.deps.Config.Storage.PhotoPath
	photoPath = utils.ExpandHomeDir(photoPath)

	filename := uuid.New().String() + filepath.Ext(fileHeader.Filename)

	// Save to disk
	if err := SaveAndResizeFile(fileHeader, filename, photoPath); err != nil {
		return nil, err
	}

	// Save metadata to DB
	folderIDKey := utils.PtrCoalesce(folderID, -1)
	keyUserPhotoList := fmt.Sprintf("user:%d:folder:%d:photo:list", userId, folderIDKey)
	keyListFirstPage := fmt.Sprintf("user:%d:folder:%d:photo:list:first-page", userId, folderIDKey)

	var photoResult *generated.Photo
	err := cache.UpdateManyAndInvalidate([]string{keyUserPhotoList, keyListFirstPage}, func() error {
		photo, err := s.repo.Create(ctx, &generated.Photo{
			UserID:     userId,
			FolderID:   folderID,
			URL:        filepath.ToSlash(filename),
			Provider:   "default",
			Name:       fileHeader.Filename,
			MetaDevice: meta["device"].(string),
			MetaOs:     meta["os"].(string),
			MetaLat:    meta["lat"].(float64),
			MetaLng:    meta["lng"].(float64),
			MetaWidth:  int(meta["width"].(float64)),
			MetaHeight: int(meta["height"].(float64)),
			MetaCapturedAt: func() *time.Time {
				if t, ok := meta["capturedAt"].(time.Time); ok {
					return &t
				}
				return nil
			}(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		})
		photoResult = photo
		return err
	})
	return photoResult, err
}

func (s *PhotoService) GetByID(ctx context.Context, id, userID int, folderID *int) (*generated.Photo, error) {
	folderIDKey := utils.PtrCoalesce(folderID, -1)
	key := fmt.Sprintf("user:%d:folder:%d:photo:id:%d", userID, folderIDKey, id)
	return cache.Get(key, cache.TTLShort, func() (*generated.Photo, error) {
		return s.repo.GetByID(ctx, id, folderID)
	})
}

func (s *PhotoService) GetByFileName(ctx context.Context, filename string, userID int, folderID *int) (*generated.Photo, error) {
	folderIDKey := utils.PtrCoalesce(folderID, -1)
	key := fmt.Sprintf("user:%d:folder:%d:photo:filename:%s", userID, folderIDKey, filename)
	return cache.Get(key, cache.TTLShort, func() (*generated.Photo, error) {
		return s.repo.GetByFileName(ctx, filename, folderID)
	})
}

func (s *PhotoService) GetAll(ctx context.Context, userID int, folderID *int) ([]*generated.Photo, error) {
	folderIDKey := utils.PtrCoalesce(folderID, -1)
	key := fmt.Sprintf("user:%d:folder:%d:photo:list", userID, folderIDKey)
	return cache.GetList(key, cache.TTLLong, func() ([]*generated.Photo, error) {
		return s.repo.GetAll(ctx, userID, folderID)
	})
}

func (s *PhotoService) BatchGetByIDs(ctx context.Context, userID int, ids []int, folderID *int) ([]*generated.Photo, error) {
	return batchUtil.BatchGetByIDs(ctx, ids, func(id int) func() (*generated.Photo, error) {
		return func() (*generated.Photo, error) {
			return s.GetByID(ctx, id, userID, folderID)
		}
	})
}

func (s *PhotoService) GetPaginated(ctx context.Context, userID int, folderID *int, page, limit int) ([]*generated.Photo, bool, error) {
	offset := (page - 1) * limit
	if page == 1 {
		folderIDKey := utils.PtrCoalesce(folderID, -1)
		keyListFirstPage := fmt.Sprintf("user:%d:folder:%d:photo:list:first-page", userID, folderIDKey)
		return cache.GetListWithHasMore(keyListFirstPage, cache.TTLLong, func() ([]*generated.Photo, bool, error) {
			return s.repo.GetPaginated(ctx, userID, folderID, limit, offset)
		})
	}
	return s.repo.GetPaginated(ctx, userID, folderID, limit, offset)
}

func (s *PhotoService) UpdateFolder(ctx context.Context, userID int, ids []int, folderID *int, oldFolderID *int) error {
	folderIDKey := utils.PtrCoalesce(folderID, -1)
	oldFolderIDKey := utils.PtrCoalesce(oldFolderID, -1)

	keysToInvalidate := []string{
		fmt.Sprintf("user:%d:folder:%d:photo:list", userID, folderIDKey),
		fmt.Sprintf("user:%d:folder:%d:photo:list:first-page", userID, folderIDKey),
		fmt.Sprintf("user:%d:folder:%d:photo:list", userID, oldFolderIDKey),
		fmt.Sprintf("user:%d:folder:%d:photo:list:first-page", userID, oldFolderIDKey),
	}

	for _, id := range ids {
		key := fmt.Sprintf("user:%d:folder:%d:photo:id:%d", userID, folderIDKey, id)
		oldKey := fmt.Sprintf("user:%d:folder:%d:photo:id:%d", userID, oldFolderIDKey, id)
		keysToInvalidate = append(keysToInvalidate, key, oldKey)
	}

	return cache.UpdateManyAndInvalidate(keysToInvalidate, func() error {
		return s.repo.UpdateFolder(ctx, ids, folderID)
	})
}

func (s *PhotoService) Delete(ctx context.Context, id, userId int, folderID *int) error {
	folderIDKey := utils.PtrCoalesce(folderID, -1)
	keyUserPhoto := fmt.Sprintf("user:%d:folder:%d:photo:id:%d", userId, folderIDKey, id)
	keyUserPhotoList := fmt.Sprintf("user:%d:folder:%d:photo:list", userId, folderIDKey)
	keyListFirstPage := fmt.Sprintf("user:%d:folder:%d:photo:list:first-page", userId, folderIDKey)
	return cache.UpdateManyAndInvalidate([]string{keyUserPhoto, keyUserPhotoList, keyListFirstPage}, func() error {
		return s.repo.SoftDelete(ctx, id)
	})
}

func (s *PhotoService) DeleteMany(ctx context.Context, ids []int, userID int, folderID *int) error {
	folderIDKey := utils.PtrCoalesce(folderID, -1)
	keyList := fmt.Sprintf("user:%d:folder:%d:photo:list", userID, folderIDKey)
	keyListFirstPage := fmt.Sprintf("user:%d:folder:%d:photo:list:first-page", userID, folderIDKey)

	var keys []string
	for _, id := range ids {
		keys = append(keys, fmt.Sprintf("user:%d:folder:%d:photo:id:%d", userID, folderIDKey, id))
	}

	keys = append(keys, keyList, keyListFirstPage)

	return cache.UpdateManyAndInvalidate(keys, func() error {
		return s.repo.SoftDeleteMany(ctx, ids)
	})
}

func (s *PhotoService) RemovePhotoFiles(photo *generated.Photo) error {
	photoPath := s.deps.Config.Storage.PhotoPath
	photoPath = utils.ExpandHomeDir(photoPath)

	sizes := []string{"original", "medium", "thumbnail"}

	var errs []error
	for _, size := range sizes {
		fullPath := filepath.Join(photoPath, size, filepath.FromSlash(photo.URL))
		if err := os.Remove(fullPath); err != nil && !os.IsNotExist(err) {
			errs = append(errs, fmt.Errorf("%s: %w", size, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to remove photo files: %v", errs)
	}
	return nil
}

func (s *PhotoService) CleanDeletedPhotoFiles(ctx context.Context) error {
	photos, err := s.repo.ListDeleted(ctx)
	if err != nil {
		return fmt.Errorf("failed to list deleted photos: %w", err)
	}
	if len(photos) == 0 {
		return nil
	}

	photoPath := utils.ExpandHomeDir(s.deps.Config.Storage.PhotoPath)
	sizes := []string{"original", "medium", "thumbnail"}

	var (
		idsToDelete []int
		errs        []error
	)

	for _, photo := range photos {
		idsToDelete = append(idsToDelete, photo.ID)

		for _, size := range sizes {
			fullPath := filepath.Join(photoPath, size, filepath.FromSlash(photo.URL))
			if err := os.Remove(fullPath); err != nil && !os.IsNotExist(err) {
				errs = append(errs, fmt.Errorf("failed to delete %s: %w", fullPath, err))
			}
		}
	}

	if err := s.repo.DeleteManyPermanently(ctx, idsToDelete); err != nil {
		errs = append(errs, fmt.Errorf("failed to hard delete from DB: %w", err))
	}

	if len(errs) > 0 {
		return fmt.Errorf("clean deleted photos: %v", errs)
	}

	return nil
}
