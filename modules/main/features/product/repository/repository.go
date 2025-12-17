package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/khiemnd777/andy_api/modules/main/config"
	model "github.com/khiemnd777/andy_api/modules/main/features/__model"
	relation "github.com/khiemnd777/andy_api/modules/main/features/__relation/policy"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/product"
	dbutils "github.com/khiemnd777/andy_api/shared/db/utils"
	"github.com/khiemnd777/andy_api/shared/logger"
	"github.com/khiemnd777/andy_api/shared/mapper"
	collectionutils "github.com/khiemnd777/andy_api/shared/metadata/collection"
	"github.com/khiemnd777/andy_api/shared/metadata/customfields"
	"github.com/khiemnd777/andy_api/shared/module"
	"github.com/khiemnd777/andy_api/shared/utils/table"
)

type ProductRepository interface {
	Create(ctx context.Context, input *model.ProductUpsertDTO) (*model.ProductDTO, error)
	Update(ctx context.Context, input *model.ProductUpsertDTO) (*model.ProductDTO, error)
	GetByID(ctx context.Context, id int) (*model.ProductDTO, error)
	List(ctx context.Context, query table.TableQuery) (table.TableListResult[model.ProductDTO], error)
	VariantList(ctx context.Context, templateID int, query table.TableQuery) (table.TableListResult[model.ProductDTO], error)
	Search(ctx context.Context, query dbutils.SearchQuery) (dbutils.SearchResult[model.ProductDTO], error)
	Delete(ctx context.Context, id int) error
}

type productRepo struct {
	db    *generated.Client
	deps  *module.ModuleDeps[config.ModuleConfig]
	cfMgr *customfields.Manager
}

func NewProductRepository(db *generated.Client, deps *module.ModuleDeps[config.ModuleConfig], cfMgr *customfields.Manager) ProductRepository {
	return &productRepo{db: db, deps: deps, cfMgr: cfMgr}
}

var productTreeCfg = collectionutils.TreeConfig{
	TableName:        "products",
	IDColumn:         "id",
	ParentIDColumn:   "template_id",
	ShowIfFieldName:  "templateId",
	CollectionGroup:  "product",
	CollectionPrefix: "product",
}

func toTreeNode(e *generated.Product) *collectionutils.TreeNode {
	return &collectionutils.TreeNode{
		ID:           e.ID,
		ParentID:     e.TemplateID,
		Name:         e.Name,
		CollectionID: e.CollectionID,
	}
}

func (r *productRepo) Create(ctx context.Context, input *model.ProductUpsertDTO) (*model.ProductDTO, error) {
	tx, err := r.db.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			_ = tx.Commit()
		}
	}()

	in := &input.DTO

	q := tx.Product.Create().
		SetNillableCode(in.Code).
		SetNillableName(in.Name).
		SetNillableCategoryID(in.CategoryID).
		SetNillableCategoryName(in.CategoryName)

	if in.TemplateID == nil {
		q.SetIsTemplate(true).
			SetNillableTemplateID(nil)
	} else {
		q.SetIsTemplate(false).
			SetNillableTemplateID(in.TemplateID)
	}

	// metadata
	if input.Collections != nil && len(*input.Collections) > 0 {
		_, err = customfields.PrepareCustomFields(ctx,
			r.cfMgr,
			*input.Collections,
			in.CustomFields,
			q,
			false,
		)
		if err != nil {
			return nil, err
		}
	}

	entity, err := q.Save(ctx)
	if err != nil {
		return nil, err
	}

	// template
	if entity.IsTemplate {
		// Upsert collection for node
		if err = collectionutils.UpsertCollectionForNode(
			ctx,
			tx,
			productTreeCfg,
			toTreeNode(entity),
			nil,
		); err != nil {
			logger.Debug(
				"product.create: upsert collection for node failed",
				"product_id", entity.ID,
				"is_template", entity.IsTemplate,
				"error", err,
			)
			return nil, err
		}

		// Upsert collections for ANCESTORS
		if err = collectionutils.UpsertAncestorCollections(
			ctx,
			tx,
			productTreeCfg,
			entity.ID,
		); err != nil {
			logger.Debug(
				"product.create: upsert ancestor collections failed",
				"product_id", entity.ID,
				"is_template", entity.IsTemplate,
				"error", err,
			)
			return nil, err
		}
	}

	out := mapper.MapAs[*generated.Product, *model.ProductDTO](entity)

	if _, err = relation.UpsertM2M(ctx, tx, "products_processes", entity, input.DTO, out); err != nil {
		return nil, err
	}

	return out, nil
}

func (r *productRepo) Update(ctx context.Context, input *model.ProductUpsertDTO) (*model.ProductDTO, error) {
	tx, err := r.db.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			_ = tx.Commit()
		}
	}()

	in := &input.DTO

	q := tx.Product.UpdateOneID(in.ID).
		SetNillableCode(in.Code).
		SetNillableName(in.Name).
		SetNillableCategoryID(in.CategoryID).
		SetNillableCategoryName(in.CategoryName)

	if in.TemplateID == nil {
		q.SetIsTemplate(true)
	} else {
		q.SetIsTemplate(false).
			SetNillableTemplateID(in.TemplateID)
	}

	// custom fields
	if input.Collections != nil && len(*input.Collections) > 0 {
		_, err = customfields.PrepareCustomFields(ctx,
			r.cfMgr,
			*input.Collections,
			in.CustomFields,
			q,
			false,
		)
		if err != nil {
			return nil, err
		}
	}

	entity, err := q.Save(ctx)
	if err != nil {
		return nil, err
	}

	// collections
	if entity.IsTemplate {
		// upsert collection for THIS NODE
		if err = collectionutils.UpsertCollectionForNode(
			ctx,
			tx,
			productTreeCfg,
			toTreeNode(entity),
			nil,
		); err != nil {
			return nil, err
		}

		// upsert collections for ANCESTORS (current branch)
		if err = collectionutils.UpsertAncestorCollections(
			ctx,
			tx,
			productTreeCfg,
			entity.ID,
		); err != nil {
			return nil, err
		}
	}

	out := mapper.MapAs[*generated.Product, *model.ProductDTO](entity)

	if _, err = relation.UpsertM2M(ctx, tx, "products_processes", entity, input.DTO, out); err != nil {
		return nil, err
	}

	return out, nil
}

func (r *productRepo) GetByID(ctx context.Context, id int) (*model.ProductDTO, error) {
	q := r.db.Product.Query().
		Where(
			product.ID(id),
			product.DeletedAtIsNil(),
		)

	entity, err := q.Only(ctx)
	if err != nil {
		return nil, err
	}

	dto := mapper.MapAs[*generated.Product, *model.ProductDTO](entity)
	return dto, nil
}

func (r *productRepo) List(ctx context.Context, query table.TableQuery) (table.TableListResult[model.ProductDTO], error) {
	list, err := table.TableList(
		ctx,
		r.db.Product.Query().
			Where(product.DeletedAtIsNil(), product.IsTemplate(true)),
		query,
		product.Table,
		product.FieldID,
		product.FieldID,
		func(src []*generated.Product) []*model.ProductDTO {
			return mapper.MapListAs[*generated.Product, *model.ProductDTO](src)
		},
	)
	if err != nil {
		var zero table.TableListResult[model.ProductDTO]
		return zero, err
	}
	return list, nil
}

func (r *productRepo) VariantList(ctx context.Context, templateID int, query table.TableQuery) (table.TableListResult[model.ProductDTO], error) {
	list, err := table.TableList(
		ctx,
		r.db.Product.Query().
			Where(product.DeletedAtIsNil(), product.TemplateIDEQ(templateID), product.IsTemplate(false)),
		query,
		product.Table,
		product.FieldID,
		product.FieldID,
		func(src []*generated.Product) []*model.ProductDTO {
			return mapper.MapListAs[*generated.Product, *model.ProductDTO](src)
		},
	)
	if err != nil {
		var zero table.TableListResult[model.ProductDTO]
		return zero, err
	}
	return list, nil
}

func (r *productRepo) Search(ctx context.Context, query dbutils.SearchQuery) (dbutils.SearchResult[model.ProductDTO], error) {
	return dbutils.Search(
		ctx,
		r.db.Product.Query().
			Where(product.DeletedAtIsNil()),
		[]string{
			dbutils.GetNormField(product.FieldCode),
			dbutils.GetNormField(product.FieldName),
		},
		query,
		product.Table,
		product.FieldID,
		product.FieldID,
		product.Or,
		func(src []*generated.Product) []*model.ProductDTO {
			return mapper.MapListAs[*generated.Product, *model.ProductDTO](src)
		},
	)
}

func (r *productRepo) Delete(ctx context.Context, id int) error {
	tx, err := r.db.Tx(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			_ = tx.Commit()
		}
	}()

	entity, err := tx.Product.Query().
		Where(
			product.IDEQ(id),
			product.DeletedAtIsNil(),
		).
		Only(ctx)
	if err != nil {
		return err
	}

	err = tx.Product.UpdateOneID(id).
		SetDeletedAt(time.Now()).
		Exec(ctx)

	if err != nil {
		return err
	}

	if entity.IsTemplate {
		if err = collectionutils.UpsertAncestorCollections(
			ctx,
			tx,
			productTreeCfg,
			id,
		); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				err = nil
			} else {
				return err
			}
		}
	}

	return nil
}
