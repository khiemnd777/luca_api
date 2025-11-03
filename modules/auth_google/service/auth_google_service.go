package service

import (
	"context"
	"fmt"
	"time"

	"github.com/khiemnd777/andy_api/modules/auth_google/config"
	"github.com/khiemnd777/andy_api/modules/auth_google/model"
	"github.com/khiemnd777/andy_api/modules/auth_google/repository"
	"github.com/khiemnd777/andy_api/shared/auth"
	"github.com/khiemnd777/andy_api/shared/module"
	tokenApi "github.com/khiemnd777/andy_api/shared/modules/token"
	"github.com/khiemnd777/andy_api/shared/pubsub"
	"github.com/khiemnd777/andy_api/shared/utils"
)

type AuthGoogleService struct {
	repo       *repository.AuthGoogleRepository
	deps       *module.ModuleDeps[config.ModuleConfig]
	secret     string
	refreshTTL time.Duration
	accessTTL  time.Duration
}

func NewAuthGoogleService(repo *repository.AuthGoogleRepository, deps *module.ModuleDeps[config.ModuleConfig], secret string) *AuthGoogleService {
	return &AuthGoogleService{
		repo:       repo,
		deps:       deps,
		secret:     secret,
		refreshTTL: 7 * 24 * time.Hour,
		accessTTL:  15 * time.Minute}
}

func (s *AuthGoogleService) LoginWithGoogleUserInfo(ctx context.Context, info *model.GoogleUserInfo) (*auth.AuthTokenPair, error) {
	user, isNew, err := s.repo.FindOrCreateUserByGoogle(ctx, &model.GoogleUserInfo{
		Email:   info.Email,
		Name:    info.Name,
		Picture: info.Picture,
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
