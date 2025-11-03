// scripts/create_module/templates/service_service.go.tmpl
package service

import (
	// "context"

	// "github.com/khiemnd777/andy_api/shared/db/ent/generated"

	"context"
	"fmt"

	"github.com/khiemnd777/andy_api/modules/notification/config"
	"github.com/khiemnd777/andy_api/modules/notification/notificationModel"
	"github.com/khiemnd777/andy_api/modules/notification/repository"
	"github.com/khiemnd777/andy_api/shared/cache"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/module"
)

type NotificationService struct {
	repo *repository.NotificationRepository
	deps *module.ModuleDeps[config.ModuleConfig]
}

func NewNotificationService(repo *repository.NotificationRepository, deps *module.ModuleDeps[config.ModuleConfig]) *NotificationService {
	return &NotificationService{
		repo: repo,
		deps: deps,
	}
}

func (s *NotificationService) LatestNotification(ctx context.Context, userID int) (*notificationModel.Notification, error) {
	return s.repo.LatestNotification(ctx, userID)
}

func (s *NotificationService) ShortListByUser(ctx context.Context, userID int) ([]*notificationModel.Notification, error) {
	key := fmt.Sprintf("user:%d:notifications:short", userID)
	return cache.GetList(key, cache.TTLLong, func() ([]*notificationModel.Notification, error) {
		return s.repo.ShortListByUser(ctx, userID)
	})
}

func (s *NotificationService) ListByUserPaginated(
	ctx context.Context,
	userID, page, limit int,
) ([]*notificationModel.Notification, bool, error) {
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	if page == 1 {
		key := fmt.Sprintf("user:%d:notifications:first-page", userID)
		return cache.GetListWithHasMore(key, cache.TTLLong, func() ([]*notificationModel.Notification, bool, error) {
			return s.repo.ListByUserPaginated(ctx, userID, limit, offset)
		})
	}

	return s.repo.ListByUserPaginated(ctx, userID, limit, offset)
}

func (s *NotificationService) GetByMessageID(ctx context.Context, messageID string) (*notificationModel.Notification, error) {
	return s.repo.GetByMessageID(ctx, messageID)
}

func (s *NotificationService) Create(ctx context.Context, messageID string, userID, notifierID int, notifType string, data map[string]any) (*generated.Notification, error) {
	var result *generated.Notification
	err := cache.UpdateManyAndInvalidate([]string{
		fmt.Sprintf("user:%d:notifications:short", userID),
		fmt.Sprintf("user:%d:notifications:first-page", userID),
		fmt.Sprintf("user:%d:notifications:unread", userID),
	}, func() error {
		notification, err := s.repo.Create(ctx, messageID, userID, notifierID, notifType, data)
		result = notification
		return err
	})
	return result, err
}

func (s *NotificationService) MarkAsRead(ctx context.Context, userID, notificationID int) error {
	return cache.UpdateManyAndInvalidate([]string{
		fmt.Sprintf("user:%d:notifications:short", userID),
		fmt.Sprintf("user:%d:notifications:first-page", userID),
		fmt.Sprintf("user:%d:notifications:unread", userID),
	}, func() error {
		return s.repo.MarkAsRead(ctx, notificationID)
	})
}

func (s *NotificationService) CountUnread(ctx context.Context, userID int) (*int, error) {
	key := fmt.Sprintf("user:%d:notifications:unread", userID)
	return cache.Get(key, cache.TTLLong, func() (*int, error) {
		return s.repo.CountUnread(ctx, userID)
	})
}

func (s *NotificationService) Delete(ctx context.Context, userID, notificationID int) error {
	return cache.UpdateManyAndInvalidate([]string{
		fmt.Sprintf("user:%d:notifications:short", userID),
		fmt.Sprintf("user:%d:notifications:first-page", userID),
		fmt.Sprintf("user:%d:notifications:unread", userID),
	}, func() error {
		return s.repo.Delete(ctx, notificationID)
	})
}

func (s *NotificationService) DeleteAll(ctx context.Context, userID int) error {
	return cache.UpdateManyAndInvalidate([]string{
		fmt.Sprintf("user:%d:notifications:short", userID),
		fmt.Sprintf("user:%d:notifications:first-page", userID),
		fmt.Sprintf("user:%d:notifications:unread", userID),
	}, func() error {
		return s.repo.DeleteAll(ctx, userID)
	})
}
