package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/khiemnd777/andy_api/modules/auth_guest/config"
	"github.com/khiemnd777/andy_api/modules/auth_guest/repository"
	"github.com/khiemnd777/andy_api/shared/auth"
	"github.com/khiemnd777/andy_api/shared/cache"
	"github.com/khiemnd777/andy_api/shared/module"
	tokenApi "github.com/khiemnd777/andy_api/shared/modules/token"
	"github.com/khiemnd777/andy_api/shared/pubsub"
	"github.com/khiemnd777/andy_api/shared/utils"
	"golang.org/x/crypto/bcrypt"
)

type AuthGuestService struct {
	repo *repository.AuthGuestRepository
	deps *module.ModuleDeps[config.ModuleConfig]
}

func NewAuthGuestService(repo *repository.AuthGuestRepository, deps *module.ModuleDeps[config.ModuleConfig]) *AuthGuestService {
	return &AuthGuestService{
		repo: repo,
		deps: deps,
	}
}

func (s *AuthGuestService) LoginWithGuest(ctx context.Context) (*auth.AuthTokenPair, error) {
	dummyAvatar := utils.GetDummyAvatarURL("Khách")
	newPassword := uuid.NewString()
	passwordhash, _ := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	user, err := s.repo.CreateNewGuest(ctx, string(passwordhash), "Khách", dummyAvatar)

	if err != nil {
		return nil, fmt.Errorf("failed to create user")
	}

	tokens, err := tokenApi.GenerateTokens(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	// Assign default role
	pubsub.PublishAsync("role:default", utils.AssignDefaultRole{
		UserID:  user.ID,
		RoleIDs: []int{3}, // Role Guest
	})

	return tokens, nil
}

func (s *AuthGuestService) DeleteGuest(ctx context.Context, userID int) error {
	keyUser := fmt.Sprintf("user:%d", userID)
	keyProfile := fmt.Sprintf("profile:%d", userID)
	return cache.UpdateManyAndInvalidate([]string{
		keyUser,
		keyProfile,
		fmt.Sprintf("user:%d:*", userID),
	}, func() error {
		return s.repo.DeleteGuest(ctx, userID)
	})
}

func (s *AuthGuestService) DeleteAllGuestsWithExpiredRefreshTokens(ctx context.Context) (int, error) {
	return s.repo.DeleteAllGuestsWithExpiredRefreshTokens(ctx)
}
