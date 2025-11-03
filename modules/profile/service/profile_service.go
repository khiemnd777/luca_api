package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/khiemnd777/andy_api/modules/profile/config"
	profileError "github.com/khiemnd777/andy_api/modules/profile/model"
	"github.com/khiemnd777/andy_api/modules/profile/repository"
	"github.com/khiemnd777/andy_api/shared/cache"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/module"
	"github.com/khiemnd777/andy_api/shared/utils"
)

type ProfileService struct {
	repo *repository.ProfileRepository
	deps *module.ModuleDeps[config.ModuleConfig]
}

func NewProfileService(repo *repository.ProfileRepository, deps *module.ModuleDeps[config.ModuleConfig]) *ProfileService {
	return &ProfileService{
		repo: repo,
		deps: deps,
	}
}

func (s *ProfileService) GetProfile(ctx context.Context, userID int) (*generated.User, error) {
	key := fmt.Sprintf("profile:%d", userID)
	return cache.Get(key, cache.TTLMedium, func() (*generated.User, error) {
		return s.repo.GetByID(ctx, userID)
	})
}

func (s *ProfileService) UpdateProfile(ctx context.Context, userID int, name, avatar string, phone, email, bankQRCode *string) (*generated.User, error) {
	// Check email
	if email != nil && *email != "" {
		if exists, _ := s.repo.CheckEmailExists(ctx, userID, *email); exists {
			return nil, profileError.ErrEmailExists
		}
	}

	// Normalize and check phone
	normalizedPhone, _ := utils.NormalizeEnsuredPhone(phone)
	if normalizedPhone != nil {
		if exists, _ := s.repo.CheckPhoneExists(ctx, userID, *normalizedPhone); exists {
			return nil, profileError.ErrPhoneExists
		}
	}

	keyUser := fmt.Sprintf("user:%d", userID)
	keyProfile := fmt.Sprintf("profile:%d", userID)

	var updated *generated.User

	err := cache.UpdateManyAndInvalidate([]string{
		keyUser,
		keyProfile,
		fmt.Sprintf("user:%d:bank_qr_code", userID),
		fmt.Sprintf("user:%d:qr_code", userID),
	}, func() error {
		var err error
		refCode := uuid.NewString()
		qrCode := utils.GenerateQRCodeStringForUser(refCode)
		updated, err = s.repo.UpdateByID(ctx, userID, name, normalizedPhone, email, &avatar, bankQRCode, &refCode, &qrCode)

		return err
	})

	return updated, err
}

func (s *ProfileService) ChangePassword(ctx context.Context, userID int, currentPassword, newPassword string) error {
	return s.repo.ChangePassword(ctx, userID, currentPassword, newPassword)
}

func (s *ProfileService) Delete(ctx context.Context, userID int) error {
	keyUser := fmt.Sprintf("user:%d", userID)
	keyProfile := fmt.Sprintf("profile:%d", userID)
	return cache.UpdateManyAndInvalidate([]string{
		keyUser,
		keyProfile,
		fmt.Sprintf("user:%d:*", userID),
	}, func() error {
		return s.repo.Delete(ctx, userID)
	})
}
