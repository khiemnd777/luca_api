package repository

import (
	"context"

	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/attribute"

	"github.com/khiemnd777/andy_api/modules/attribute/config"
	"github.com/khiemnd777/andy_api/shared/module"
)

type AttributeRepository struct {
	client *generated.Client
	deps   *module.ModuleDeps[config.ModuleConfig]
}

func NewAttributeRepository(client *generated.Client, deps *module.ModuleDeps[config.ModuleConfig]) *AttributeRepository {
	return &AttributeRepository{
		client: client,
		deps:   deps,
	}
}

func (r *AttributeRepository) Create(ctx context.Context, data *generated.Attribute) (*generated.Attribute, error) {
	return r.client.Attribute.Create().
		SetUserID(data.UserID).
		SetAttributeName(data.AttributeName).
		SetAttributeType(data.AttributeType).
		SetDeleted(false).
		Save(ctx)
}

func (r *AttributeRepository) GetByID(ctx context.Context, id int) (*generated.Attribute, error) {
	return r.client.Attribute.Get(ctx, id)
}

func (r *AttributeRepository) ListByUser(ctx context.Context, userID int) ([]*generated.Attribute, error) {
	return r.client.Attribute.
		Query().
		Where(attribute.UserID(userID), attribute.Deleted(false)).
		Order(generated.Desc(attribute.FieldUpdatedAt)).
		All(ctx)
}

func (r *AttributeRepository) ListByUserPaginated(ctx context.Context, userID, limit, offset int) ([]*generated.Attribute, bool, error) {
	items, err := r.client.Attribute.
		Query().
		Where(attribute.UserID(userID), attribute.Deleted(false)).
		Order(generated.Desc(attribute.FieldUpdatedAt)).
		Limit(limit + 1).
		Offset(offset).
		All(ctx)
	if err != nil {
		return nil, false, err
	}

	hasMore := len(items) > limit
	if hasMore {
		items = items[:limit]
	}
	return items, hasMore, nil
}

func (r *AttributeRepository) Update(ctx context.Context, id int, name, attributeType string) (*generated.Attribute, error) {
	return r.client.Attribute.UpdateOneID(id).
		SetAttributeName(name).
		SetAttributeType(attributeType).
		Save(ctx)
}

func (r *AttributeRepository) SoftDelete(ctx context.Context, id int) error {
	_, err := r.client.Attribute.UpdateOneID(id).SetDeleted(true).Save(ctx)
	return err
}
