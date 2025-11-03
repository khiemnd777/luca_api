package repository

import (
	"context"
	"time"

	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/refreshtoken"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/user"
)

type AuthRepository struct {
	db *generated.Client
}

func NewAuthRepository(db *generated.Client) *AuthRepository {
	return &AuthRepository{db: db}
}

func (r *AuthRepository) GetUserByEmail(ctx context.Context, email string) (*generated.User, error) {
	return r.db.User.Query().Where(user.Email(email)).Only(ctx)
}

func (r *AuthRepository) GetUserByPhone(ctx context.Context, phone string) (*generated.User, error) {
	return r.db.User.Query().Where(user.Phone(phone)).Only(ctx)
}

func (r *AuthRepository) CreateNewUser(ctx context.Context, phone, email *string, password, name string, avatar *string, refCode *string, qrCode *string) (*generated.User, error) {
	return r.db.User.Create().
		SetNillablePhone(phone).
		SetNillableEmail(email).
		SetName(name).
		SetPassword(password).
		SetNillableAvatar(avatar).
		SetNillableQrCode(qrCode).
		SetNillableRefCode(refCode).
		SetProvider("system").
		Save(ctx)
}

func (r *AuthRepository) CheckPhoneExists(ctx context.Context, phone string) (bool, error) {
	return r.db.User.Query().
		Where(user.PhoneEQ(phone)).
		Exist(ctx)
}

func (r *AuthRepository) CheckEmailExists(ctx context.Context, email string) (bool, error) {
	return r.db.User.Query().
		Where(user.EmailEQ(email)).
		Exist(ctx)
}

func (r *AuthRepository) CreateRefreshToken(ctx context.Context, userID int, token string, expiresAt time.Time) error {
	_, err := r.db.RefreshToken.Create().
		SetUserID(userID).
		SetToken(token).
		SetExpiresAt(expiresAt).
		Save(ctx)
	return err
}

func (r *AuthRepository) DeleteRefreshToken(ctx context.Context, token string) error {
	_, err := r.db.RefreshToken.Delete().
		Where(refreshtoken.Token(token)).
		Exec(ctx)
	return err
}

func (r *AuthRepository) IsRefreshTokenValid(ctx context.Context, token string) (bool, int, string, error) {
	t, err := r.db.RefreshToken.Query().
		Where(refreshtoken.TokenEQ(token)).
		WithUser().
		Only(ctx)
	if err != nil {
		return false, 0, "", err
	}
	if time.Now().After(t.ExpiresAt) {
		return false, 0, "", nil
	}
	return true, t.Edges.User.ID, t.Edges.User.Email, nil
}
