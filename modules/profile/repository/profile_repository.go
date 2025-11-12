package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/khiemnd777/andy_api/modules/profile/config"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/user"
	"github.com/khiemnd777/andy_api/shared/module"
	"golang.org/x/crypto/bcrypt"
)

type ProfileRepository struct {
	db   *generated.Client
	deps *module.ModuleDeps[config.ModuleConfig]
}

func NewProfileRepository(db *generated.Client, deps *module.ModuleDeps[config.ModuleConfig]) *ProfileRepository {
	return &ProfileRepository{
		db:   db,
		deps: deps,
	}
}

func (r *ProfileRepository) GetByID(ctx context.Context, id int) (*generated.User, error) {
	return r.db.User.Query().Where(user.ID(id), user.DeletedAtIsNil()).Only(ctx)
}

func (r *ProfileRepository) UpdateByID(ctx context.Context, id int, name string, phone, email, avatar *string, refCode *string, qrCode *string) (*generated.User, error) {
	tx, err := r.db.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	u, err := tx.User.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	upd := tx.User.UpdateOneID(id).
		SetName(name).
		SetNillablePhone(phone).
		SetNillableEmail(email).
		SetNillableAvatar(avatar)

	if u.QrCode == nil && refCode != nil && qrCode != nil {
		upd = upd.SetRefCode(*refCode).SetQrCode(*qrCode)
	}

	res, err := upd.Save(ctx)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return res, nil
}

func (r *ProfileRepository) ChangePassword(ctx context.Context, id int, currentPassword, newPassword string) error {
	var currentHash string
	const selectQuery = `SELECT password FROM users WHERE id = $1`
	err := r.deps.DB.QueryRowContext(ctx, selectQuery, id).Scan(&currentHash)
	if err != nil {
		return fmt.Errorf("failed to get current password hash: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(currentHash), []byte(currentPassword)); err != nil {
		return fmt.Errorf("current password is incorrect")
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	const updateQuery = `UPDATE users SET password = $2 WHERE id = $1`
	_, err = r.deps.DB.ExecContext(ctx, updateQuery, id, string(newHash))
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

func (r *ProfileRepository) CheckPhoneExists(ctx context.Context, userID int, phone string) (bool, error) {
	return r.db.User.Query().
		Where(user.IDNEQ(userID), user.PhoneEQ(phone), user.DeletedAtIsNil()).
		Exist(ctx)
}

func (r *ProfileRepository) CheckEmailExists(ctx context.Context, userID int, email string) (bool, error) {
	return r.db.User.Query().
		Where(user.IDNEQ(userID), user.EmailEQ(email), user.DeletedAtIsNil()).
		Exist(ctx)
}

func (r *ProfileRepository) Delete(ctx context.Context, id int) error {
	_, err := r.db.User.
		UpdateOneID(id).
		SetActive(false).
		SetDeletedAt(time.Now()).
		Save(ctx)
	return err
}
