package repository

import (
	"context"

	"github.com/khiemnd777/andy_api/modules/user/model"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/role"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/user"
)

type UserRepository struct {
	db *generated.Client
}

func NewUserRepository(db *generated.Client) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, email, password, name string, phone, avatar *string, bankQRCode *string, refCode *string, qrCode *string) (*generated.User, error) {
	return r.db.User.Create().
		SetEmail(email).
		SetPassword(password).
		SetName(name).
		SetNillablePhone(phone).
		SetNillableAvatar(avatar).
		SetNillableRefCode(refCode).
		SetNillableQrCode(qrCode).
		Save(ctx)
}

func (r *UserRepository) GetByID(ctx context.Context, id int) (*generated.User, error) {
	return r.db.User.Query().Where(user.ID(id), user.Active(true)).Only(ctx)
}

func (r *UserRepository) GetAdminUserID(ctx context.Context) (*int, error) {
	adminUser, err := r.db.User.
		Query().
		Where(user.HasRolesWith(role.RoleNameEQ("admin"))).
		First(ctx)
	if err != nil {
		return nil, err
	}
	if adminUser == nil {
		return nil, nil
	}
	return &adminUser.ID, nil
}

func (r *UserRepository) GetQRCodeByUserID(ctx context.Context, userID int) (*model.QRCodeModel, error) {
	user, err := r.db.User.Query().Where(user.ID(userID), user.Active(true)).Only(ctx)
	if err != nil {
		return nil, err
	}
	if user.QrCode == nil {
		return nil, nil
	}
	result := &model.QRCodeModel{
		ID:     user.ID,
		QRCode: *user.QrCode,
		Name:   user.Name,
		Avatar: &user.Avatar,
	}
	return result, nil
}

func (r *UserRepository) GeUserByRefCode(ctx context.Context, refCode string) (*generated.User, error) {
	user, err := r.db.User.Query().Where(user.RefCode(refCode), user.Active(true)).Only(ctx)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *UserRepository) BatchGetByIDs(ctx context.Context, ids []int) (map[int]*generated.User, error) {
	if len(ids) == 0 {
		return map[int]*generated.User{}, nil
	}

	users, err := r.db.User.
		Query().
		Where(user.IDIn(ids...)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	userMap := make(map[int]*generated.User, len(users))
	for _, u := range users {
		userMap[u.ID] = u
	}

	return userMap, nil
}

func (r *UserRepository) Update(ctx context.Context, id int, name string, phone, avatar *string, bankQRCode *string, refCode *string, qrCode *string) (*generated.User, error) {
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

func (r *UserRepository) Delete(ctx context.Context, id int) error {
	return r.db.User.DeleteOneID(id).Exec(ctx)
}
