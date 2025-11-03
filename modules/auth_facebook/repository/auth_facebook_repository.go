package repository

import (
	"context"

	"github.com/khiemnd777/andy_api/modules/auth_facebook/config"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/user"
	"github.com/khiemnd777/andy_api/shared/module"
	"github.com/khiemnd777/andy_api/shared/utils"
)

type FacebookUserInfo struct {
	Email   string
	Name    string
	Picture string
	Sub     string // Facebook user ID
}

type AuthFacebookRepository struct {
	ent  *generated.Client
	deps *module.ModuleDeps[config.ModuleConfig]
}

func NewAuthFacebookRepository(ent *generated.Client, deps *module.ModuleDeps[config.ModuleConfig]) *AuthFacebookRepository {
	return &AuthFacebookRepository{
		ent:  ent,
		deps: deps,
	}
}

func (r *AuthFacebookRepository) FindOrCreateUserByFacebook(ctx context.Context, info *FacebookUserInfo) (*generated.User, bool, error) {
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
		SetProvider("facebook").
		SetProviderID(info.Sub).
		Save(ctx)

	if err != nil {
		return nil, false, err
	}

	return new, true, nil
}
