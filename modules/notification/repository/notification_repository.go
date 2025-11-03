// scripts/create_module/templates/repository_repo.go.tmpl
package repository

import (
	"context"

	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/notification"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/user"

	"github.com/khiemnd777/andy_api/modules/notification/config"
	"github.com/khiemnd777/andy_api/modules/notification/notificationModel"
	"github.com/khiemnd777/andy_api/shared/module"
)

type NotificationRepository struct {
	client *generated.Client
	deps   *module.ModuleDeps[config.ModuleConfig]
}

func NewNotificationRepository(client *generated.Client, deps *module.ModuleDeps[config.ModuleConfig]) *NotificationRepository {
	return &NotificationRepository{
		client: client,
		deps:   deps,
	}
}

func (r *NotificationRepository) LatestNotification(ctx context.Context, userID int) (*notificationModel.Notification, error) {
	single, err := r.client.Notification.Query().
		Where(
			notification.UserID(userID),
			notification.Read(false),
			notification.Deleted(false),
		).
		Order(generated.Desc(notification.FieldCreatedAt)).
		First(ctx)

	if err != nil {
		return nil, err
	}

	result := &notificationModel.Notification{
		ID:         single.ID,
		UserID:     single.UserID,
		NotifierID: single.NotifierID,
		CreatedAt:  single.CreatedAt,
		Type:       single.Type,
		Read:       single.Read,
		Data:       single.Data,
	}

	if notifier, err := r.client.User.
		Query().
		Where(user.ID(single.NotifierID)).
		Only(ctx); err == nil {
		result.Notifier = notifier
	}

	return result, nil
}

func (r *NotificationRepository) GetByMessageID(ctx context.Context, messageID string) (*notificationModel.Notification, error) {
	single, err := r.client.Notification.Query().
		Where(
			notification.MessageID(messageID),
			notification.Read(false),
			notification.Deleted(false),
		).
		Only(ctx)

	if err != nil {
		return nil, err
	}

	result := &notificationModel.Notification{
		ID:         single.ID,
		UserID:     single.UserID,
		NotifierID: single.NotifierID,
		CreatedAt:  single.CreatedAt,
		Type:       single.Type,
		Read:       single.Read,
		Data:       single.Data,
	}

	if notifier, err := r.client.User.
		Query().
		Where(user.ID(single.NotifierID)).
		Only(ctx); err == nil {
		result.Notifier = notifier
	}

	return result, nil
}

func (r *NotificationRepository) ShortListByUser(ctx context.Context, userID int) ([]*notificationModel.Notification, error) {
	notifs, err := r.client.Notification.
		Query().
		Where(notification.UserIDEQ(userID), notification.Deleted(false)).
		Order(generated.Desc(notification.FieldCreatedAt)).
		Limit(7).
		All(ctx)

	if err != nil {
		return nil, err
	}

	var result []*notificationModel.Notification

	for _, n := range notifs {
		nElm := notificationModel.Notification{
			ID:         n.ID,
			UserID:     n.UserID,
			NotifierID: n.NotifierID,
			CreatedAt:  n.CreatedAt,
			Type:       n.Type,
			Read:       n.Read,
			Data:       n.Data,
		}
		notifier, err := r.client.User.
			Query().
			Where(user.ID(n.NotifierID)).
			Only(ctx)
		if err == nil {
			nElm.Notifier = notifier
		}

		result = append(result, &nElm)
	}

	return result, nil
}

func (r *NotificationRepository) ListByUserPaginated(ctx context.Context, userID, limit, offset int) ([]*notificationModel.Notification, bool, error) {
	notifs, err := r.client.Notification.
		Query().
		Where(notification.UserIDEQ(userID), notification.Deleted(false)).
		Order(generated.Desc(notification.FieldCreatedAt)).
		Offset(offset).
		Limit(limit + 1).
		All(ctx)

	if err != nil {
		return nil, false, err
	}

	hasMore := len(notifs) > limit
	if hasMore {
		notifs = notifs[:limit]
	}

	var result []*notificationModel.Notification

	for _, n := range notifs {
		nElm := notificationModel.Notification{
			ID:         n.ID,
			UserID:     n.UserID,
			NotifierID: n.NotifierID,
			CreatedAt:  n.CreatedAt,
			Type:       n.Type,
			Read:       n.Read,
			Data:       n.Data,
		}
		notifier, err := r.client.User.
			Query().
			Where(user.ID(n.NotifierID)).
			Only(ctx)
		if err == nil {
			nElm.Notifier = notifier
		}

		result = append(result, &nElm)
	}

	return result, hasMore, nil
}

func (r *NotificationRepository) Create(ctx context.Context, messageID string, userID, notifierID int, notifType string, data map[string]any) (*generated.Notification, error) {
	return r.client.Notification.
		Create().
		SetUserID(userID).
		SetNotifierID(notifierID).
		SetMessageID(messageID).
		SetType(notifType).
		SetRead(false).
		SetData(data).
		Save(ctx)
}

func (r *NotificationRepository) MarkAsRead(ctx context.Context, notificationID int) error {
	return r.client.Notification.
		UpdateOneID(notificationID).
		SetRead(true).
		Exec(ctx)
}

func (r *NotificationRepository) CountUnread(ctx context.Context, userID int) (*int, error) {
	count, err := r.client.Notification.
		Query().
		Where(
			notification.UserIDEQ(userID),
			notification.Read(false),
			notification.Deleted(false),
		).
		Count(ctx)

	if err != nil {
		return nil, err
	}

	return &count, nil
}

func (r *NotificationRepository) Delete(ctx context.Context, notificationID int) error {
	return r.client.Notification.
		UpdateOneID(notificationID).
		SetDeleted(true).
		Exec(ctx)
}

func (r *NotificationRepository) DeleteAll(ctx context.Context, userID int) error {
	_, err := r.client.Notification.
		Update().
		Where(
			notification.UserID(userID),
			notification.Deleted(false),
		).
		SetDeleted(true).
		Save(ctx)
	return err
}
