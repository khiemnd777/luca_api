// scripts/create_module/templates/repository_repo.go.tmpl
package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/photo"
	"github.com/khiemnd777/andy_api/shared/logger"

	"github.com/khiemnd777/andy_api/modules/photo/config"
	"github.com/khiemnd777/andy_api/shared/module"
)

type PhotoRepository struct {
	db   *generated.Client
	deps *module.ModuleDeps[config.ModuleConfig]
}

func NewPhotoRepository(db *generated.Client, deps *module.ModuleDeps[config.ModuleConfig]) *PhotoRepository {
	return &PhotoRepository{
		db:   db,
		deps: deps,
	}
}

func (r *PhotoRepository) Create(ctx context.Context, input *generated.Photo) (*generated.Photo, error) {
	created, err := r.db.Photo.Create().
		SetUserID(input.UserID).
		SetNillableFolderID(input.FolderID).
		SetURL(input.URL).
		SetProvider(input.Provider).
		SetName(input.Name).
		SetMetaDevice(input.MetaDevice).
		SetMetaOs(input.MetaOs).
		SetMetaLat(input.MetaLat).
		SetMetaLng(input.MetaLng).
		SetMetaWidth(input.MetaWidth).
		SetMetaHeight(input.MetaHeight).
		SetNillableMetaCapturedAt(input.MetaCapturedAt).
		Save(ctx)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to create photo: %v", err))
		return nil, err
	}
	return created, nil
}

func (r *PhotoRepository) UpdateFolder(ctx context.Context, photoIDs []int, folderID *int) error {
	if len(photoIDs) == 0 {
		return nil
	}
	_, err := r.db.Photo.
		Update().
		Where(photo.IDIn(photoIDs...), photo.Deleted(false)).
		SetNillableFolderID(folderID).
		Save(ctx)
	if err != nil {
		logger.Error("Failed to update photo folder: ", err)
	}
	return err
}

func (r *PhotoRepository) GetByID(ctx context.Context, id int, folderID *int) (*generated.Photo, error) {
	query := r.db.Photo.Query().
		Where(photo.ID(id), photo.Deleted(false))

	if folderID != nil {
		query = query.Where(photo.FolderID(*folderID))
	}

	return query.Only(ctx)
}

func (r *PhotoRepository) GetByFileName(ctx context.Context, filename string, folderID *int) (*generated.Photo, error) {
	query := r.db.Photo.Query().
		Where(photo.URL(filename), photo.Deleted(false))

	if folderID != nil {
		query = query.Where(photo.FolderID(*folderID))
	}

	return query.Only(ctx)
}

func (r *PhotoRepository) GetAll(ctx context.Context, userID int, folderID *int) ([]*generated.Photo, error) {
	query := r.db.Photo.Query().
		Where(photo.UserID(userID), photo.Deleted(false)).
		Order(generated.Desc(photo.FieldUpdatedAt))

	if folderID != nil {
		if *folderID == -1 {
			query = query.Where(photo.FolderIDIsNil())
		} else {
			query = query.Where(photo.FolderID(*folderID))
		}
	}

	return query.All(ctx)
}

func (r *PhotoRepository) GetPaginated(ctx context.Context, userID int, folderID *int, limit, offset int) ([]*generated.Photo, bool, error) {
	query := r.db.Photo.
		Query().
		Where(photo.UserID(userID), photo.Deleted(false))

	if folderID != nil {
		if *folderID == -1 {
			query = query.Where(photo.FolderIDIsNil())
		} else {
			query = query.Where(photo.FolderID(*folderID))
		}
	}

	var results []struct {
		ID             int        `json:"id"`
		Name           string     `json:"name"`
		Url            string     `json:"url"`
		MetaCapturedAt *time.Time `json:"meta_captured_at,omitempty"`
		CreatedAt      time.Time  `json:"created_at"`
	}

	err := query.
		Order(generated.Desc(photo.FieldUpdatedAt)).
		Limit(limit+1).
		Offset(offset).
		Select(photo.FieldID, photo.FieldName, photo.FieldURL, photo.FieldMetaCapturedAt, photo.FieldCreatedAt).
		Scan(ctx, &results)

	if err != nil {
		return nil, false, err
	}

	hasMore := len(results) > limit
	if hasMore {
		results = results[:limit]
	}

	photos := make([]*generated.Photo, len(results))
	for i, p := range results {
		photos[i] = &generated.Photo{
			ID:             p.ID,
			Name:           p.Name,
			URL:            p.Url,
			MetaCapturedAt: p.MetaCapturedAt,
			CreatedAt:      p.CreatedAt,
		}
	}

	return photos, hasMore, nil
}

func (r *PhotoRepository) SoftDelete(ctx context.Context, id int) error {
	return r.db.Photo.UpdateOneID(id).
		SetDeleted(true).
		Exec(ctx)
}

func (r *PhotoRepository) SoftDeleteMany(ctx context.Context, ids []int) error {
	if len(ids) == 0 {
		return nil
	}
	_, err := r.db.Photo.
		Update().
		Where(photo.IDIn(ids...), photo.Deleted(false)).
		SetDeleted(true).
		Save(ctx)
	return err
}

func (r *PhotoRepository) DeletePermanently(ctx context.Context, id int) error {
	return r.db.Photo.DeleteOneID(id).Exec(ctx)
}

func (r *PhotoRepository) DeleteManyPermanently(ctx context.Context, ids []int) error {
	if len(ids) == 0 {
		return nil
	}
	_, err := r.db.Photo.
		Delete().
		Where(photo.IDIn(ids...)).
		Exec(ctx)
	return err
}

func (r *PhotoRepository) ListDeleted(ctx context.Context) ([]*generated.Photo, error) {
	return r.db.Photo.
		Query().
		Where(photo.Deleted(true)).
		All(ctx)
}
