package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/khiemnd777/andy_api/modules/auth_guest/config"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/refreshtoken"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/role"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/user"
	"github.com/khiemnd777/andy_api/shared/module"
)

type AuthGuestRepository struct {
	db   *generated.Client
	deps *module.ModuleDeps[config.ModuleConfig]
}

func NewAuthGuestRepository(db *generated.Client, deps *module.ModuleDeps[config.ModuleConfig]) *AuthGuestRepository {
	return &AuthGuestRepository{
		db:   db,
		deps: deps,
	}
}

func (r *AuthGuestRepository) CreateNewGuest(ctx context.Context, password, name, avatar string) (*generated.User, error) {
	return r.db.User.Create().
		SetPassword(password).
		SetName(name).
		SetAvatar(avatar).
		SetProvider("system").
		Save(ctx)
}

func (r *AuthGuestRepository) DeleteGuest(ctx context.Context, userID int) error {
	tx, err := r.db.Tx(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = commitOrRollback(tx, &err) }()

	// 1) Verify role 'guest'
	hasGuest, err := tx.User.
		Query().
		Where(
			user.ID(userID),
			user.HasRolesWith(role.RoleNameEqualFold("guest")),
		).
		Exist(ctx)
	if err != nil {
		return err
	}
	if !hasGuest {
		return fmt.Errorf("user %d is not a guest", userID)
	}

	// 2) Delete refresh tokens (điều chỉnh field/edge cho đúng schema)
	if _, err = tx.RefreshToken.
		Delete().
		Where(refreshtoken.HasUserWith(user.ID(userID))).
		Exec(ctx); err != nil {
		return err
	}

	// 3) Clear roles ở bảng pivot user_roles
	if err = tx.User.
		UpdateOneID(userID).
		ClearRoles().
		Exec(ctx); err != nil {
		return err
	}

	// 4) Delete user
	if err = tx.User.
		DeleteOneID(userID).
		Exec(ctx); err != nil {
		return err
	}

	return nil
}

func commitOrRollback(tx *generated.Tx, pErr *error) error {
	if *pErr != nil {
		_ = tx.Rollback()
		return *pErr
	}
	return tx.Commit()
}

func (r *AuthGuestRepository) DeleteAllGuestsWithExpiredRefreshTokens(ctx context.Context) (int, error) {
	now := time.Now()

	// 1) Lấy danh sách ID user là guest và có refresh token đã hết hạn
	ids, err := r.db.User.
		Query().
		Where(
			user.HasRolesWith(role.RoleNameEqualFold("guest")),
			user.HasRefreshTokensWith(refreshtoken.ExpiresAtLT(now)),
		).
		IDs(ctx)
	if err != nil {
		return 0, err
	}
	if len(ids) == 0 {
		return 0, nil
	}

	// 2) Transaction xoá theo thứ tự: refresh_tokens -> clear roles -> users
	tx, err := r.db.Tx(ctx)
	if err != nil {
		return 0, err
	}
	defer func() { _ = commitOrRollback(tx, &err) }()

	// 2.1) Xoá refresh tokens (edge predicate)
	if _, err = tx.RefreshToken.
		Delete().
		Where(refreshtoken.HasUserWith(user.IDIn(ids...))).
		Exec(ctx); err != nil {
		return 0, err
	}

	// 2.2) Clear roles (pivot user_roles) cho tất cả user target
	if err = tx.User.
		Update().
		Where(user.IDIn(ids...)).
		ClearRoles().
		Exec(ctx); err != nil {
		return 0, err
	}

	// 2.3) Xoá users
	if _, err = tx.User.
		Delete().
		Where(user.IDIn(ids...)).
		Exec(ctx); err != nil {
		return 0, err
	}

	return len(ids), nil
}
