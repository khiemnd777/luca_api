// scripts/create_module/templates/repository_repo.go.tmpl
package repository

import (
	"context"

	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/folder"

	"github.com/khiemnd777/andy_api/modules/folder/config"
	"github.com/khiemnd777/andy_api/shared/module"
)

type FolderRepository struct {
	client *generated.Client
	deps   *module.ModuleDeps[config.ModuleConfig]
}

func NewFolderRepository(client *generated.Client, deps *module.ModuleDeps[config.ModuleConfig]) *FolderRepository {
	return &FolderRepository{
		client: client,
		deps:   deps,
	}
}

func (r *FolderRepository) Create(ctx context.Context, input *generated.Folder) (*generated.Folder, error) {
	return r.client.Folder.Create().
		SetUserID(input.UserID).
		SetFolderName(input.FolderName).
		SetColor(input.Color).
		SetShared(input.Shared).
		SetNillableParentID(input.ParentID).
		Save(ctx)
}

func (r *FolderRepository) FindByID(ctx context.Context, id int) (*generated.Folder, error) {
	return r.client.Folder.Get(ctx, id)
}

func (r *FolderRepository) ListByUser(ctx context.Context, userID int) ([]*generated.Folder, error) {
	return r.client.Folder.
		Query().
		Where(
			folder.UserID(userID),
			folder.Deleted(false),
		).
		Order(generated.Desc(folder.FieldUpdatedAt)).
		All(ctx)
}

func (r *FolderRepository) ListByUserPaginated(ctx context.Context, userID, limit, offset int) ([]*generated.Folder, bool, error) {
	items, err := r.client.Folder.
		Query().
		Where(folder.UserID(userID), folder.Deleted(false)).
		Order(generated.Desc(folder.FieldUpdatedAt)).
		Limit(limit + 1).
		Offset(offset).
		All(ctx)
	if err != nil {
		return nil, false, err
	}

	hasMore := len(items) > limit
	if hasMore {
		items = items[:limit]
	}
	return items, hasMore, nil
}

func (r *FolderRepository) Update(ctx context.Context, id int, input *generated.Folder) (*generated.Folder, error) {
	return r.client.Folder.UpdateOneID(id).
		SetFolderName(input.FolderName).
		SetColor(input.Color).
		SetShared(input.Shared).
		SetNillableParentID(input.ParentID).
		Save(ctx)
}

func (r *FolderRepository) SoftDelete(ctx context.Context, id int) error {
	_, err := r.client.Folder.UpdateOneID(id).SetDeleted(true).Save(ctx)
	return err
}
