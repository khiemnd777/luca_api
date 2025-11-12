package repository

import (
	"context"

	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/role"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/user"
	dbutils "github.com/khiemnd777/andy_api/shared/db/utils"
	"github.com/khiemnd777/andy_api/shared/utils/table"
)

type RoleRepository interface {
	Create(ctx context.Context, name, displayName, brief string) (*generated.Role, error)
	GetByID(ctx context.Context, id int) (*generated.Role, error)
	GetByName(ctx context.Context, name string) (*generated.Role, error)
	GetAll(ctx context.Context) ([]*generated.Role, error)
	List(ctx context.Context, query table.TableQuery) (table.TableListResult[generated.Role], error)
	ListByUser(ctx context.Context, userID, limit, offset int) ([]*generated.Role, int, error)
	ListByUserID(ctx context.Context, userID int, query table.TableQuery) (table.TableListResult[generated.Role], error)
	SearchRoles(ctx context.Context, q dbutils.SearchQuery) (dbutils.SearchResult[generated.Role], error)
	Update(ctx context.Context, id int, newName, newDisplayName, newBrief string) (*generated.Role, error)
	UpdateName(ctx context.Context, id int, newName string) (*generated.Role, error)
	Delete(ctx context.Context, id int) error

	// Matrix ops
	ReplacePermissions(ctx context.Context, roleID int, permIDs []int) error
	AddPermissions(ctx context.Context, roleID int, permIDs []int) error
	RemovePermissions(ctx context.Context, roleID int, permIDs []int) error

	// Helper
	UserIDsOfRole(ctx context.Context, roleID int) ([]int, error)
	PermissionIDsOfRole(ctx context.Context, roleID int) ([]int, error)
}

type roleRepository struct{ db *generated.Client }

func NewRoleRepository(db *generated.Client) RoleRepository { return &roleRepository{db: db} }

func (r *roleRepository) Create(ctx context.Context, name, displayName, brief string) (*generated.Role, error) {
	return r.db.Role.Create().
		SetRoleName(name).
		SetDisplayName(displayName).
		SetBrief(brief).
		Save(ctx)
}
func (r *roleRepository) GetByID(ctx context.Context, id int) (*generated.Role, error) {
	return r.db.Role.Get(ctx, id)
}
func (r *roleRepository) GetByName(ctx context.Context, name string) (*generated.Role, error) {
	return r.db.Role.Query().Where(role.RoleNameEQ(name)).First(ctx)
}
func (r *roleRepository) GetAll(ctx context.Context) ([]*generated.Role, error) {
	return r.db.Role.Query().Order(generated.Asc(role.FieldID)).All(ctx)
}
func (r *roleRepository) List(ctx context.Context, query table.TableQuery) (table.TableListResult[generated.Role], error) {
	return table.TableList[generated.Role, generated.Role](
		ctx,
		r.db.Role.Query(),
		query,
		role.Table,
		role.FieldID,
		role.FieldID,
		nil,
	)
}
func (r *roleRepository) ListByUser(ctx context.Context, userID, limit, offset int) ([]*generated.Role, int, error) {
	q := r.db.Role.Query().Where(role.HasUsersWith(user.IDEQ(userID)))
	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	items, err := q.Limit(limit).Offset(offset).Order(generated.Desc(role.FieldID)).All(ctx)
	return items, total, err
}
func (r *roleRepository) ListByUserID(ctx context.Context, userID int, query table.TableQuery) (table.TableListResult[generated.Role], error) {
	return table.TableList[generated.Role, generated.Role](
		ctx,
		r.db.Role.Query().
			Where(role.HasUsersWith((user.IDEQ(userID)))),
		query,
		role.Table,
		role.FieldID,
		role.FieldID,
		nil,
	)
}

func (r *roleRepository) SearchRoles(ctx context.Context, query dbutils.SearchQuery) (dbutils.SearchResult[generated.Role], error) {
	return dbutils.Search[generated.Role, generated.Role](
		ctx,
		r.db.Role.Query(),
		[]string{
			dbutils.GetNormField(role.FieldDisplayName),
			dbutils.GetNormField(role.FieldRoleName),
		},
		query,
		role.Table,
		role.FieldID,
		role.FieldRoleName,
		role.Or,
		nil,
	)
}

func (r *roleRepository) UpdateName(ctx context.Context, id int, newName string) (*generated.Role, error) {
	return r.db.Role.UpdateOneID(id).SetRoleName(newName).Save(ctx)
}

func (r *roleRepository) Update(ctx context.Context, id int, newName, newDisplayName, newBrief string) (*generated.Role, error) {
	return r.db.Role.UpdateOneID(id).
		SetRoleName(newName).
		SetDisplayName(newDisplayName).
		SetBrief(newBrief).
		Save(ctx)
}
func (r *roleRepository) Delete(ctx context.Context, id int) error {
	// Clear edges to avoid FK issues then delete
	tx, err := r.db.Tx(ctx)
	if err != nil {
		return err
	}
	if err := tx.Role.UpdateOneID(id).ClearPermissions().ClearUsers().Exec(ctx); err != nil {
		_ = tx.Rollback()
		return err
	}
	if err := tx.Role.DeleteOneID(id).Exec(ctx); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

func (r *roleRepository) ReplacePermissions(ctx context.Context, roleID int, permIDs []int) error {
	return r.db.Role.UpdateOneID(roleID).ClearPermissions().AddPermissionIDs(permIDs...).Exec(ctx)
}
func (r *roleRepository) AddPermissions(ctx context.Context, roleID int, permIDs []int) error {
	return r.db.Role.UpdateOneID(roleID).AddPermissionIDs(permIDs...).Exec(ctx)
}
func (r *roleRepository) RemovePermissions(ctx context.Context, roleID int, permIDs []int) error {
	return r.db.Role.UpdateOneID(roleID).RemovePermissionIDs(permIDs...).Exec(ctx)
}

func (r *roleRepository) UserIDsOfRole(ctx context.Context, roleID int) ([]int, error) {
	users, err := r.db.Role.Query().Where(role.IDEQ(roleID)).QueryUsers().All(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]int, 0, len(users))
	for _, u := range users {
		out = append(out, u.ID)
	}
	return out, nil
}
func (r *roleRepository) PermissionIDsOfRole(ctx context.Context, roleID int) ([]int, error) {
	perms, err := r.db.Role.Query().Where(role.IDEQ(roleID)).QueryPermissions().All(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]int, 0, len(perms))
	for _, p := range perms {
		out = append(out, p.ID)
	}
	return out, nil
}
