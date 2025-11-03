package service

import (
	"context"
	"fmt"
	"time"

	"github.com/khiemnd777/andy_api/modules/auth_apple/config"
	"github.com/khiemnd777/andy_api/modules/auth_apple/model"
	"github.com/khiemnd777/andy_api/modules/auth_apple/repository"
	"github.com/khiemnd777/andy_api/shared/auth"
	"github.com/khiemnd777/andy_api/shared/module"
	tokenApi "github.com/khiemnd777/andy_api/shared/modules/token"
	"github.com/khiemnd777/andy_api/shared/pubsub"
	"github.com/khiemnd777/andy_api/shared/utils"
)

type AuthAppleService struct {
	repo       *repository.AuthAppleRepository
	deps       *module.ModuleDeps[config.ModuleConfig]
	secret     string
	refreshTTL time.Duration
	accessTTL  time.Duration
}

func NewAuthAppleService(repo *repository.AuthAppleRepository, deps *module.ModuleDeps[config.ModuleConfig], secret string) *AuthAppleService {
	return &AuthAppleService{
		repo:       repo,
		deps:       deps,
		secret:     secret,
		refreshTTL: 7 * 24 * time.Hour,
		accessTTL:  15 * time.Minute}
}

func (s *AuthAppleService) LoginWithAppleUserInfo(ctx context.Context, info *model.AppleUserInfo) (*auth.AuthTokenPair, error) {
	dummyAvatar := utils.GetDummyAvatarURL(info.Name)

	user, isNew, err := s.repo.FindOrCreateUserByApple(ctx, &model.AppleUserInfo{
		Email:   info.Email,
		Name:    info.Name,
		Picture: dummyAvatar,
		Sub:     info.Sub,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create user")
	}

	tokens, err := tokenApi.GenerateTokens(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	if isNew {
		// Assign default role
		pubsub.PublishAsync("role:default", utils.AssignDefaultRole{
			UserID:  user.ID,
			RoleIDs: []int{1},
		})
	}

	return tokens, nil
}
