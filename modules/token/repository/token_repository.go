package repository

import (
	"context"
	"strings"
	"time"

	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/department"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/departmentmember"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/refreshtoken"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/user"
)

type TokenRepository struct {
	db *generated.Client
}

func NewTokenRepository(db *generated.Client) *TokenRepository {
	return &TokenRepository{db: db}
}

func (r *TokenRepository) GetUserByID(ctx context.Context, id int) (*generated.User, error) {
	return r.db.User.Query().Where(user.ID(id)).Only(ctx)
}

func (r *TokenRepository) GetPermissionsByUserID(ctx context.Context, id int) (*map[string]struct{}, error) {
	perms, dbErr := r.db.User.
		Query().
		Where(user.IDEQ(id)).
		QueryRoles().
		QueryPermissions().
		All(ctx)
	if dbErr != nil {
		return nil, dbErr
	}
	set := make(map[string]struct{}, len(perms))
	for _, p := range perms {
		if p == nil {
			continue
		}
		val := strings.ToLower(strings.TrimSpace(p.PermissionValue))
		if val != "" {
			set[val] = struct{}{}
		}
	}
	return &set, nil
}

func (r *TokenRepository) GetDepartmentByUserID(ctx context.Context, id int) (*int, error) {
	dm, err := r.db.DepartmentMember.
		Query().
		Where(departmentmember.UserID(id)).
		Order(departmentmember.ByCreatedAt()).
		First(ctx)

	if err != nil {
		return nil, err
	}

	dept, err := dm.QueryDepartment().
		Where(department.Deleted(false)).
		Select(department.FieldID).
		First(ctx)
	if err != nil {
		return nil, err
	}

	return &dept.ID, nil
}

func (r *TokenRepository) GetUserByEmail(ctx context.Context, email string) (*generated.User, error) {
	return r.db.User.Query().Where(user.Email(email)).Only(ctx)
}

func (r *TokenRepository) CreateRefreshToken(ctx context.Context, userID int, token string, expiresAt time.Time) error {
	_, err := r.db.RefreshToken.Create().
		SetUserID(userID).
		SetToken(token).
		SetExpiresAt(expiresAt).
		Save(ctx)
	return err
}

func (r *TokenRepository) DeleteRefreshToken(ctx context.Context, token string) error {
	_, err := r.db.RefreshToken.Delete().
		Where(refreshtoken.Token(token)).
		Exec(ctx)
	return err
}

func (r *TokenRepository) IsRefreshTokenValid(ctx context.Context, tok string) (found bool, valid bool, userID int, email string, err error) {
	t, err := r.db.RefreshToken.
		Query().
		Where(refreshtoken.TokenEQ(tok)).
		WithUser(func(uq *generated.UserQuery) {
			uq.Select(user.FieldID, user.FieldEmail)
		}).
		Only(ctx)

	if err != nil {
		if generated.IsNotFound(err) {
			return false, false, 0, "", nil // not found
		}
		return false, false, 0, "", err // other DB error
	}

	if time.Now().After(t.ExpiresAt) {
		return true, false, t.Edges.User.ID, t.Edges.User.Email, nil // found but expired
	}

	return true, true, t.Edges.User.ID, t.Edges.User.Email, nil // valid
}

func (r *TokenRepository) IsRefreshTokenValid2(ctx context.Context, token string) (bool, int, string, error) {
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

func (r *TokenRepository) DeleteExpiredRefreshTokens(ctx context.Context) error {
	_, err := r.db.RefreshToken.Delete().
		Where(refreshtoken.ExpiresAtLT(time.Now())).
		Exec(ctx)
	return err
}
