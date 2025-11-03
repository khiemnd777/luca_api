package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/khiemnd777/andy_api/modules/user/config"
	"github.com/khiemnd777/andy_api/modules/user/model"
	"github.com/khiemnd777/andy_api/modules/user/repository"
	batchUtil "github.com/khiemnd777/andy_api/shared/batch"
	"github.com/khiemnd777/andy_api/shared/cache"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/module"
	"github.com/khiemnd777/andy_api/shared/utils"
)

type UserService struct {
	repo *repository.UserRepository
	deps *module.ModuleDeps[config.ModuleConfig]
}

func NewUserService(repo *repository.UserRepository, deps *module.ModuleDeps[config.ModuleConfig]) *UserService {
	return &UserService{repo: repo, deps: deps}
}

func (s *UserService) Create(ctx context.Context, email, password, name string, phone *string) (*generated.User, error) {
	normalizedPhone := utils.NormalizePhone(phone)
	dummyAvatar := utils.GetDummyAvatarURL(name)
	refCode := uuid.NewString()
	qrCode := utils.GenerateQRCodeStringForUser(refCode)
	return s.repo.Create(ctx, email, password, name, &normalizedPhone, &dummyAvatar, nil, &refCode, &qrCode)
}

func (s *UserService) GetByID(ctx context.Context, id int) (*generated.User, error) {
	key := fmt.Sprintf("user:%d", id)
	return cache.Get(key, cache.TTLLong, func() (*generated.User, error) {
		return s.repo.GetByID(ctx, id)
	})
}

func (s *UserService) GetAdminUserID(ctx context.Context) (*int, error) {
	return cache.Get("user:admin", cache.TTLLong, func() (*int, error) {
		return s.repo.GetAdminUserID(ctx)
	})
}

func (s *UserService) GetQRCodeByUserID(ctx context.Context, userID int) (*model.QRCodeModel, error) {
	key := fmt.Sprintf("user:%d:qr_code", userID)
	return cache.Get(key, cache.TTLLong, func() (*model.QRCodeModel, error) {
		return s.repo.GetQRCodeByUserID(ctx, userID)
	})
}

func (s *UserService) GetUserByRefCode(ctx context.Context, refCode string) (*generated.User, error) {
	key := fmt.Sprintf("user:ref:%s", refCode)
	return cache.Get(key, cache.TTLLong, func() (*generated.User, error) {
		return s.repo.GeUserByRefCode(ctx, refCode)
	})
}

func (s *UserService) BatchGetByIDs(ctx context.Context, ids []int) ([]*generated.User, error) {
	return batchUtil.BatchGetByIDs(ctx, ids, func(id int) func() (*generated.User, error) {
		return func() (*generated.User, error) {
			return s.GetByID(ctx, id)
		}
	})
}

func (s *UserService) Update(ctx context.Context, id int, name string, phone, avatar *string, bankQRCode *string) (*generated.User, error) {
	keyUser := fmt.Sprintf("user:%d", id)
	keyProfile := fmt.Sprintf("profile:%d", id)
	var updated *generated.User
	err := cache.UpdateManyAndInvalidate([]string{
		keyUser,
		keyProfile,
		fmt.Sprintf("user:%d:bank_qr_code", id),
		fmt.Sprintf("user:%d:qr_code", id),
	}, func() error {
		var err error
		normalizedPhone := utils.NormalizePhone(phone)
		var consumedAvatar *string
		if avatar == nil {
			consumedAvatar = utils.Ptr(utils.GetDummyAvatarURL(name))
		} else {
			consumedAvatar = avatar
		}

		refCode := uuid.NewString()
		qrCode := utils.GenerateQRCodeStringForUser(refCode)

		updated, err = s.repo.Update(ctx, id, name, &normalizedPhone, consumedAvatar, bankQRCode, &refCode, &qrCode)
		return err
	})
	return updated, err
}

func (s *UserService) Delete(ctx context.Context, id int) error {
	keyUser := fmt.Sprintf("user:%d", id)
	keyProfile := fmt.Sprintf("profile:%d", id)
	return cache.UpdateManyAndInvalidate([]string{
		keyUser,
		keyProfile,
		fmt.Sprintf("user:%d:*", id),
	}, func() error {
		return s.repo.Delete(ctx, id)
	})
}
