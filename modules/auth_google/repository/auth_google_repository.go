package repository

import (
	"context"

	"github.com/khiemnd777/andy_api/modules/auth_google/config"
	"github.com/khiemnd777/andy_api/modules/auth_google/model"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/user"
	"github.com/khiemnd777/andy_api/shared/module"
	"github.com/khiemnd777/andy_api/shared/utils"
)

type AuthGoogleRepository struct {
	ent  *generated.Client
	deps *module.ModuleDeps[config.ModuleConfig]
}

func NewAuthGoogleRepository(ent *generated.Client, deps *module.ModuleDeps[config.ModuleConfig]) *AuthGoogleRepository {
	return &AuthGoogleRepository{
		ent:  ent,
		deps: deps,
	}
}

func (r *AuthGoogleRepository) FindOrCreateUserByGoogle(ctx context.Context, info *model.GoogleUserInfo) (*generated.User, bool, error) {
	u, err := r.ent.User.Query().Where(user.ProviderIDEQ(info.Sub), user.Active(true)).First(ctx)

	if err == nil {
		return u, false, nil
	}

	dummyPassword := utils.GenerateOAuthDummyPassword()

	new, err := r.ent.User.Create().
		SetEmail(info.Email).
		SetName(info.Name).
		SetAvatar(info.Picture).
		SetPassword(dummyPassword).
		SetProvider("google").
		SetProviderID(info.Sub).
		Save(ctx)

	if err != nil {
		return nil, false, err
	}

	return new, true, nil
}
