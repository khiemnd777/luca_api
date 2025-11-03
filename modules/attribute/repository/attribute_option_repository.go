package repository

import (
	"context"

	"github.com/khiemnd777/andy_api/modules/attribute/config"
	attribute "github.com/khiemnd777/andy_api/modules/attribute/model"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/attributeoption"
	"github.com/khiemnd777/andy_api/shared/module"
)

type AttributeOptionRepository struct {
	client *generated.Client
	deps   *module.ModuleDeps[config.ModuleConfig]
}

func NewAttributeOptionRepository(client *generated.Client, deps *module.ModuleDeps[config.ModuleConfig]) *AttributeOptionRepository {
	return &AttributeOptionRepository{
		client: client,
		deps:   deps,
	}
}

func (r *AttributeOptionRepository) Create(ctx context.Context, option *generated.AttributeOption) (*generated.AttributeOption, error) {
	return r.client.AttributeOption.Create().
		SetUserID(option.UserID).
		SetAttributeID(option.AttributeID).
		SetOptionValue(option.OptionValue).
		SetDisplayOrder(option.DisplayOrder).
		Save(ctx)
}

func (r *AttributeOptionRepository) Update(ctx context.Context, id int, value string, order int) (*generated.AttributeOption, error) {
	return r.client.AttributeOption.UpdateOneID(id).
		SetOptionValue(value).
		SetDisplayOrder(order).
		Save(ctx)
}

func (r *AttributeOptionRepository) SoftDelete(ctx context.Context, id int) error {
	_, err := r.client.AttributeOption.UpdateOneID(id).
		SetDeleted(true).
		Save(ctx)
	return err
}

func (r *AttributeOptionRepository) ListByAttribute(ctx context.Context, attributeID int) ([]*generated.AttributeOption, error) {
	return r.client.AttributeOption.
		Query().
		Where(attributeoption.AttributeID(attributeID), attributeoption.Deleted(false)).
		Order(generated.Asc(attributeoption.FieldDisplayOrder)).
		All(ctx)
}

func (r *AttributeOptionRepository) GetByID(ctx context.Context, id int) (*generated.AttributeOption, error) {
	return r.client.AttributeOption.Get(ctx, id)
}

func (r *AttributeOptionRepository) BatchUpdateDisplayOrder(
	ctx context.Context,
	orders []attribute.OptionOrder,
) error {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return err
	}
	for _, o := range orders {
		if err := tx.AttributeOption.UpdateOneID(o.OptionID).
			SetDisplayOrder(o.DisplayOrder).
			Exec(ctx); err != nil {
			_ = tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}
