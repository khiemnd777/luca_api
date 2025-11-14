// modules/permission/repository/permission_repository.go
package repository

import (
	"context"

	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/permission"
)

type PermissionRepository interface {
	Create(ctx context.Context, name, value string) (*generated.Permission, error)
	GetByID(ctx context.Context, id int) (*generated.Permission, error)
	GetByValue(ctx context.Context, value string) (*generated.Permission, error)
	List(ctx context.Context, limit, offset int) ([]*generated.Permission, int, error)
	Delete(ctx context.Context, id int) error
}

type permissionRepository struct{ db *generated.Client }

func NewPermissionRepository(db *generated.Client) PermissionRepository {
	return &permissionRepository{db: db}
}

func (r *permissionRepository) Create(ctx context.Context, name, value string) (*generated.Permission, error) {
	return r.db.Permission.Create().SetPermissionName(name).SetPermissionValue(value).Save(ctx)
}
func (r *permissionRepository) GetByID(ctx context.Context, id int) (*generated.Permission, error) {
	return r.db.Permission.Get(ctx, id)
}
func (r *permissionRepository) GetByValue(ctx context.Context, value string) (*generated.Permission, error) {
	return r.db.Permission.Query().Where(permission.PermissionValueEQ(value)).First(ctx)
}
func (r *permissionRepository) List(ctx context.Context, limit, offset int) ([]*generated.Permission, int, error) {
	q := r.db.Permission.Query()
	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	items, err := q.Limit(limit).Offset(offset).Order(generated.Asc(permission.FieldPermissionValue)).All(ctx)
	return items, total, err
}
func (r *permissionRepository) Delete(ctx context.Context, id int) error {
	// Clear edges then delete
	tx, err := r.db.Tx(ctx)
	if err != nil {
		return err
	}
	if err := tx.Permission.UpdateOneID(id).ClearRoles().Exec(ctx); err != nil {
		_ = tx.Rollback()
		return err
	}
	if err := tx.Permission.DeleteOneID(id).Exec(ctx); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}
