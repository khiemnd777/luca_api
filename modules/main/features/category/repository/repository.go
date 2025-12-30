package repository

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/khiemnd777/andy_api/modules/main/config"
	model "github.com/khiemnd777/andy_api/modules/main/features/__model"
	relation "github.com/khiemnd777/andy_api/modules/main/features/__relation/policy"
	"github.com/khiemnd777/andy_api/modules/metadata/repository"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/category"
	dbutils "github.com/khiemnd777/andy_api/shared/db/utils"
	"github.com/khiemnd777/andy_api/shared/logger"
	"github.com/khiemnd777/andy_api/shared/mapper"
	collectionutils "github.com/khiemnd777/andy_api/shared/metadata/collection"
	"github.com/khiemnd777/andy_api/shared/metadata/customfields"
	"github.com/khiemnd777/andy_api/shared/module"
	"github.com/khiemnd777/andy_api/shared/utils"
	"github.com/khiemnd777/andy_api/shared/utils/table"
)

type CategoryRepository interface {
	Create(ctx context.Context, input *model.CategoryUpsertDTO) (*model.CategoryDTO, error)
	Update(ctx context.Context, input *model.CategoryUpsertDTO) (*model.CategoryDTO, error)
	GetByID(ctx context.Context, id int) (*model.CategoryDTO, error)
	List(ctx context.Context, query table.TableQuery) (table.TableListResult[model.CategoryDTO], error)
	Search(ctx context.Context, query dbutils.SearchQuery) (dbutils.SearchResult[model.CategoryDTO], error)
	Delete(ctx context.Context, id int) error
}

type categoryRepo struct {
	db             *generated.Client
	deps           *module.ModuleDeps[config.ModuleConfig]
	cfMgr          *customfields.Manager
	collectionRepo *repository.CollectionRepository
}

func NewCategoryRepository(db *generated.Client, deps *module.ModuleDeps[config.ModuleConfig], cfMgr *customfields.Manager) CategoryRepository {
	collectionRepo := repository.NewCollectionRepository(deps.DB)
	return &categoryRepo{
		db:             db,
		deps:           deps,
		cfMgr:          cfMgr,
		collectionRepo: collectionRepo,
	}
}

var categoryTreeCfg = collectionutils.TreeConfig{
	TableName:        "categories",
	IDColumn:         "id",
	ParentIDColumn:   "parent_id",
	ShowIfFieldName:  "categoryId",
	CollectionGroup:  "category",
	CollectionPrefix: "category",
}

func toTreeNode(e *generated.Category) *collectionutils.TreeNode {
	return &collectionutils.TreeNode{
		ID:           e.ID,
		ParentID:     e.ParentID,
		Name:         e.Name,
		CollectionID: e.CollectionID,
	}
}

func (r *categoryRepo) Create(ctx context.Context, input *model.CategoryUpsertDTO) (*model.CategoryDTO, error) {
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

	dto := &input.DTO

	q := tx.Category.Create().
		SetNillableName(dto.Name).
		SetNillableParentID(dto.ParentID).
		SetNillableCategoryIDLv1(dto.CategoryIDLv1).
		SetNillableCategoryNameLv1(dto.CategoryNameLv1).
		SetNillableCategoryIDLv2(dto.CategoryIDLv2).
		SetNillableCategoryNameLv2(dto.CategoryNameLv2).
		SetNillableCategoryIDLv3(dto.CategoryIDLv3).
		SetNillableCategoryNameLv3(dto.CategoryNameLv3)

	if dto.Level > 0 {
		q.SetLevel(dto.Level)
	}

	if input.Collections != nil && len(*input.Collections) > 0 {
		_, err = customfields.PrepareCustomFields(ctx,
			r.cfMgr,
			*input.Collections,
			dto.CustomFields,
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
	// Upsert collection for node
	if err = collectionutils.UpsertCollectionForNode(
		ctx,
		tx,
		categoryTreeCfg,
		toTreeNode(entity),
		nil,
	); err != nil {
		return nil, err
	}

	// Upsert collections for ANCESTORS
	if err = collectionutils.UpsertAncestorCollections(
		ctx,
		tx,
		categoryTreeCfg,
		entity.ID,
	); err != nil {
		return nil, err
	}

	out := mapper.MapAs[*generated.Category, *model.CategoryDTO](entity)

	if _, err = relation.UpsertM2M(ctx, tx, "categories_processes", entity, input.DTO, out); err != nil {
		return nil, err
	}

	return out, nil
}

func (r *categoryRepo) Update(ctx context.Context, input *model.CategoryUpsertDTO) (*model.CategoryDTO, error) {
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

	dto := &input.DTO

	prevCategory, err := tx.Category.Query().
		Where(
			category.ID(dto.ID),
			category.DeletedAtIsNil(),
		).
		Only(ctx)
	if err != nil {
		return nil, err
	}

	q := tx.Category.UpdateOneID(dto.ID).
		SetNillableName(dto.Name).
		SetNillableParentID(dto.ParentID).
		SetNillableCategoryIDLv1(dto.CategoryIDLv1).
		SetNillableCategoryNameLv1(dto.CategoryNameLv1).
		SetNillableCategoryIDLv2(dto.CategoryIDLv2).
		SetNillableCategoryNameLv2(dto.CategoryNameLv2).
		SetNillableCategoryIDLv3(dto.CategoryIDLv3).
		SetNillableCategoryNameLv3(dto.CategoryNameLv3)

	parentLevel := 0
	if dto.ParentID != nil {
		parent, err := tx.Category.Query().
			Where(
				category.ID(*dto.ParentID),
				category.DeletedAtIsNil(),
			).
			Only(ctx)
		if err != nil {
			return nil, err
		}
		parentLevel = parent.Level
	}

	if parentLevel < 3 {
		q.ClearCategoryIDLv3().
			ClearCategoryNameLv3()
	}
	if parentLevel < 2 {
		q.ClearCategoryIDLv2().
			ClearCategoryNameLv2()
	}
	if parentLevel < 1 {
		q.ClearCategoryIDLv1().
			ClearCategoryNameLv1()
	}

	if dto.Level > 0 {
		q.SetLevel(dto.Level)
	}

	if input.Collections != nil && len(*input.Collections) > 0 {
		_, err = customfields.PrepareCustomFields(
			ctx,
			r.cfMgr,
			*input.Collections,
			dto.CustomFields,
			q,
			true,
		)
		if err != nil {
			return nil, err
		}
	}

	entity, err := q.Save(ctx)
	if err != nil {
		return nil, err
	}

	parentChanged := (prevCategory.ParentID == nil && entity.ParentID != nil) ||
		(prevCategory.ParentID != nil && (entity.ParentID == nil || *entity.ParentID != *prevCategory.ParentID))

	if parentChanged {
		if err = r.updateDescendantsLineage(ctx, tx, entity); err != nil {
			return nil, err
		}
	}

	// collections
	// upsert collection for THIS NODE
	if err = collectionutils.UpsertCollectionForNode(
		ctx,
		tx,
		categoryTreeCfg,
		toTreeNode(entity),
		nil,
	); err != nil {
		return nil, err
	}

	// upsert collections for ANCESTORS (current branch)
	if err = collectionutils.UpsertAncestorCollections(
		ctx,
		tx,
		categoryTreeCfg,
		entity.ID,
	); err != nil {
		return nil, err
	}

	// if parent changed → upsert ancestors of OLD branch
	if parentChanged && prevCategory.Level > 0 && prevCategory.ParentID != nil {
		if err = collectionutils.UpsertAncestorCollections(
			ctx,
			tx,
			categoryTreeCfg,
			*prevCategory.ParentID,
		); err != nil {
			return nil, err
		}
	}

	out := mapper.MapAs[*generated.Category, *model.CategoryDTO](entity)

	if _, err = relation.UpsertM2M(ctx, tx, "categories_processes", entity, input.DTO, out); err != nil {
		return nil, err
	}

	return out, nil
}

func (r *categoryRepo) GetByID(ctx context.Context, id int) (*model.CategoryDTO, error) {
	q := r.db.Category.Query().
		Where(
			category.ID(id),
			category.DeletedAtIsNil(),
		)

	entity, err := q.Only(ctx)
	if err != nil {
		return nil, err
	}

	dto := mapper.MapAs[*generated.Category, *model.CategoryDTO](entity)
	return dto, nil
}

func (r *categoryRepo) List(ctx context.Context, query table.TableQuery) (table.TableListResult[model.CategoryDTO], error) {
	list, err := table.TableList(
		ctx,
		r.db.Category.Query().
			Where(category.DeletedAtIsNil()),
		query,
		category.Table,
		category.FieldID,
		category.FieldID,
		func(src []*generated.Category) []*model.CategoryDTO {
			ordered := orderCategoriesTree(src)
			return mapper.MapListAs[*generated.Category, *model.CategoryDTO](ordered)
		},
	)
	if err != nil {
		var zero table.TableListResult[model.CategoryDTO]
		return zero, err
	}
	return list, nil
}

func (r *categoryRepo) Search(ctx context.Context, query dbutils.SearchQuery) (dbutils.SearchResult[model.CategoryDTO], error) {
	return dbutils.Search(
		ctx,
		r.db.Category.Query().
			Where(category.DeletedAtIsNil()),
		[]string{
			dbutils.GetNormField(category.FieldName),
		},
		query,
		category.Table,
		category.FieldID,
		category.FieldID,
		category.Or,
		func(src []*generated.Category) []*model.CategoryDTO {
			ordered := orderCategoriesTree(src)
			return mapper.MapListAs[*generated.Category, *model.CategoryDTO](ordered)
		},
	)
}

func (r *categoryRepo) Delete2(ctx context.Context, id int) error {
	return r.db.Category.UpdateOneID(id).
		SetDeletedAt(time.Now()).
		Exec(ctx)
}

func (r *categoryRepo) Delete(ctx context.Context, id int) (err error) {
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

	entity, err := tx.Category.Query().
		Where(
			category.ID(id),
			category.DeletedAtIsNil(),
		).
		Only(ctx)
	if err != nil {
		return err
	}

	logger.Debug(fmt.Sprintf("Deleting category: %+v", entity))

	countChildren, err := tx.Category.Query().
		Where(
			category.ParentID(id),
			category.DeletedAtIsNil(),
		).
		Count(ctx)
	if err != nil {
		return err
	}
	if countChildren > 0 {
		return fmt.Errorf("cannot delete category %d because it still has child categories", id)
	}

	if err = collectionutils.UpsertAncestorCollections(
		ctx,
		tx,
		categoryTreeCfg,
		entity.ID,
	); err != nil {
		return err
	}

	if err = tx.Category.UpdateOneID(id).
		SetDeletedAt(time.Now()).
		Exec(ctx); err != nil {
		return err
	}

	logger.Debug(fmt.Sprintf("Deleting category: %+v", err))

	return nil
}

// -- helpers

type categoryLineage struct {
	level   int
	lv1ID   *int
	lv1Name *string
	lv2ID   *int
	lv2Name *string
	lv3ID   *int
	lv3Name *string
}

func (r *categoryRepo) updateDescendantsLineage(ctx context.Context, tx *generated.Tx, entity *generated.Category) error {
	ids, err := collectionutils.CollectDescendantIDs(
		ctx,
		tx,
		categoryTreeCfg,
		entity.ID,
	)
	if err != nil {
		return err
	}
	if len(ids) == 0 {
		return nil
	}

	descendants, err := tx.Category.Query().
		Where(
			category.IDIn(ids...),
			category.DeletedAtIsNil(),
		).
		All(ctx)
	if err != nil {
		return err
	}

	childrenByParent := make(map[int][]*generated.Category, len(descendants))
	for _, child := range descendants {
		if child.ParentID == nil {
			continue
		}
		childrenByParent[*child.ParentID] = append(childrenByParent[*child.ParentID], child)
	}

	baseLineage, err := r.buildCategoryLineage(ctx, tx, entity)
	if err != nil {
		return err
	}

	queue := []struct {
		parent  *generated.Category
		lineage categoryLineage
	}{{
		parent:  entity,
		lineage: baseLineage,
	}}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		children := childrenByParent[current.parent.ID]
		for _, child := range children {
			childLineage := deriveChildLineage(current.parent, current.lineage, child)
			_, err = tx.Category.UpdateOneID(child.ID).
				SetLevel(childLineage.level).
				SetNillableCategoryIDLv1(childLineage.lv1ID).
				SetNillableCategoryNameLv1(childLineage.lv1Name).
				SetNillableCategoryIDLv2(childLineage.lv2ID).
				SetNillableCategoryNameLv2(childLineage.lv2Name).
				SetNillableCategoryIDLv3(childLineage.lv3ID).
				SetNillableCategoryNameLv3(childLineage.lv3Name).
				Save(ctx)
			if err != nil {
				return err
			}

			queue = append(queue, struct {
				parent  *generated.Category
				lineage categoryLineage
			}{
				parent:  child,
				lineage: childLineage,
			})
		}
	}

	return nil
}

func (r *categoryRepo) buildCategoryLineage(ctx context.Context, tx *generated.Tx, entity *generated.Category) (categoryLineage, error) {
	if entity.ParentID == nil {
		return categoryLineage{
			level: 1,
		}, nil
	}

	parent, err := tx.Category.Query().
		Where(
			category.ID(*entity.ParentID),
			category.DeletedAtIsNil(),
		).
		Only(ctx)
	if err != nil {
		return categoryLineage{}, err
	}

	parentLineage, err := r.buildCategoryLineage(ctx, tx, parent)
	if err != nil {
		return categoryLineage{}, err
	}

	return deriveChildLineage(parent, parentLineage, entity), nil
}

func deriveChildLineage(parent *generated.Category, parentLineage categoryLineage, child *generated.Category) categoryLineage {
	childLevel := parentLineage.level + 1

	type ancestor struct {
		id   *int
		name *string
	}

	ancestors := make([]ancestor, 0, 3)
	if parentLineage.lv1ID != nil {
		ancestors = append(ancestors, ancestor{parentLineage.lv1ID, parentLineage.lv1Name})
	}
	if parentLineage.lv2ID != nil {
		ancestors = append(ancestors, ancestor{parentLineage.lv2ID, parentLineage.lv2Name})
	}
	if parentLineage.lv3ID != nil {
		ancestors = append(ancestors, ancestor{parentLineage.lv3ID, parentLineage.lv3Name})
	}
	ancestors = append(ancestors, ancestor{utils.Ptr(parent.ID), parent.Name})

	var lv1ID, lv2ID, lv3ID *int
	var lv1Name, lv2Name, lv3Name *string

	if len(ancestors) > 0 {
		lv1ID = ancestors[0].id
		lv1Name = ancestors[0].name
	}
	if len(ancestors) > 1 {
		lv2ID = ancestors[1].id
		lv2Name = ancestors[1].name
	}
	if len(ancestors) > 2 {
		lv3ID = ancestors[2].id
		lv3Name = ancestors[2].name
	}

	return categoryLineage{
		level:   childLevel,
		lv1ID:   lv1ID,
		lv1Name: lv1Name,
		lv2ID:   lv2ID,
		lv2Name: lv2Name,
		lv3ID:   lv3ID,
		lv3Name: lv3Name,
	}
}

func orderCategoriesTree(entities []*generated.Category) []*generated.Category {
	childrenByParent := make(map[int][]*generated.Category, len(entities))
	var roots []*generated.Category

	for _, entity := range entities {
		if entity.ParentID == nil {
			roots = append(roots, entity)
			continue
		}
		parentID := *entity.ParentID
		childrenByParent[parentID] = append(childrenByParent[parentID], entity)
	}

	// sort level 0 and each children group
	sort.Slice(roots, func(i, j int) bool { return roots[i].ID < roots[j].ID })
	for pid, children := range childrenByParent {
		sort.Slice(children, func(i, j int) bool { return children[i].ID < children[j].ID })
		childrenByParent[pid] = children
	}

	ordered := make([]*generated.Category, 0, len(entities))
	visited := make(map[int]struct{}, len(entities))

	var walk func(nodes []*generated.Category)
	walk = func(nodes []*generated.Category) {
		for _, node := range nodes {
			if _, ok := visited[node.ID]; ok {
				continue
			}
			visited[node.ID] = struct{}{}
			ordered = append(ordered, node)
			if children, ok := childrenByParent[node.ID]; ok {
				walk(children)
			}
		}
	}

	walk(roots)

	// fallback: in case some nodes have missing parents in the filtered result
	if len(ordered) < len(entities) {
		for _, entity := range entities {
			if _, ok := visited[entity.ID]; ok {
				continue
			}
			ordered = append(ordered, entity)
			if children, ok := childrenByParent[entity.ID]; ok {
				walk(children)
			}
		}
	}

	return ordered
}
