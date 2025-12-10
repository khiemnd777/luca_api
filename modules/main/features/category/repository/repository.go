package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/khiemnd777/andy_api/modules/main/config"
	model "github.com/khiemnd777/andy_api/modules/main/features/__model"
	relation "github.com/khiemnd777/andy_api/modules/main/features/__relation/policy"
	"github.com/khiemnd777/andy_api/modules/metadata/repository"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/category"
	dbutils "github.com/khiemnd777/andy_api/shared/db/utils"
	"github.com/khiemnd777/andy_api/shared/mapper"
	"github.com/khiemnd777/andy_api/shared/metadata/customfields"
	"github.com/khiemnd777/andy_api/shared/module"
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

func (r *categoryRepo) upsertCollection(ctx context.Context, tx *generated.Tx, entity *generated.Category, conds []customfields.ShowIfCondition) (*generated.Category, error) {
	if len(conds) == 0 {
		conds = []customfields.ShowIfCondition{{
			Field: "categoryId",
			Op:    "eq",
			Value: entity.ID,
		}}
	}
	showIf := customfields.ShowIfCondition{Any: conds}

	showIfJSON, err := json.Marshal(showIf)
	if err != nil {
		return nil, err
	}

	collectionSlug := "category-" + strconv.Itoa(entity.ID)
	collectionName := "Category"
	if entity.Name != nil && *entity.Name != "" {
		collectionName = *entity.Name
	}

	collectionGroup := "category"
	showIfValue := sql.NullString{String: string(showIfJSON), Valid: true}

	if entity.CollectionID != nil {
		integration := true
		_, err = r.collectionRepo.Update(ctx, *entity.CollectionID, &collectionSlug, &collectionName, &showIfValue, &integration, &collectionGroup)
		if err != nil {
			return nil, err
		}
		return entity, nil
	}

	collection, err := r.collectionRepo.Create(ctx, collectionSlug, collectionName, &showIfValue, true, &collectionGroup)
	if err != nil {
		return nil, err
	}

	return tx.Category.UpdateOneID(entity.ID).
		SetCollectionID(collection.ID).
		Save(ctx)
}

func (r *categoryRepo) collectDescendantIDs(ctx context.Context, tx *generated.Tx, parentID int) ([]int, error) {
	queue := []int{parentID}
	seen := map[int]struct{}{parentID: {}}
	var ids []int

	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]

		children, err := tx.Category.Query().
			Where(
				category.ParentID(id),
				category.DeletedAtIsNil(),
			).
			All(ctx)
		if err != nil {
			return nil, err
		}

		for _, child := range children {
			if _, ok := seen[child.ID]; ok {
				continue
			}
			seen[child.ID] = struct{}{}
			ids = append(ids, child.ID)
			queue = append(queue, child.ID)
		}
	}

	return ids, nil
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

	entity, err = r.upsertCollection(ctx, tx, entity, nil)
	if err != nil {
		return nil, err
	}

	if err = r.upsertAncestorCollections(ctx, tx, entity); err != nil {
		return nil, err
	}

	dto = mapper.MapAs[*generated.Category, *model.CategoryDTO](entity)

	err = relation.Upsert1(ctx, tx, "category", entity, &input.DTO, dto)
	if err != nil {
		return nil, err
	}

	_, err = relation.UpsertM2M(ctx, tx, "category", entity, input.DTO, dto)
	if err != nil {
		return nil, err
	}

	return dto, nil
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

	entity, err = r.upsertCollection(ctx, tx, entity, nil)
	if err != nil {
		return nil, err
	}

	if err = r.upsertAncestorCollections(ctx, tx, entity); err != nil {
		return nil, err
	}

	if prevCategory.Level > 0 && prevCategory.ParentID != nil && (entity.ParentID == nil || *entity.ParentID != *prevCategory.ParentID) {
		prevParentID := *prevCategory.ParentID
		if err = r.upsertAncestorCollections(ctx, tx, &generated.Category{ParentID: &prevParentID}); err != nil {
			return nil, err
		}
	}

	dto = mapper.MapAs[*generated.Category, *model.CategoryDTO](entity)

	err = relation.Upsert1(ctx, tx, "category", entity, &input.DTO, dto)
	if err != nil {
		return nil, err
	}

	_, err = relation.UpsertM2M(ctx, tx, "category", entity, input.DTO, dto)
	if err != nil {
		return nil, err
	}

	return dto, nil
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

func (r *categoryRepo) Delete(ctx context.Context, id int) error {
	return r.db.Category.UpdateOneID(id).
		SetDeletedAt(time.Now()).
		Exec(ctx)
}

// -- helpers

func (r *categoryRepo) upsertAncestorCollections(ctx context.Context, tx *generated.Tx, entity *generated.Category) error {
	if entity.ParentID == nil {
		return nil
	}

	visited := map[int]bool{}
	parentID := entity.ParentID

	for parentID != nil {
		if visited[*parentID] {
			return fmt.Errorf("category cycle detected at ID %d", *parentID)
		}
		visited[*parentID] = true

		parent, err := tx.Category.Query().
			Where(
				category.ID(*parentID),
				category.DeletedAtIsNil(),
			).
			Only(ctx)
		if err != nil {
			return err
		}

		descendants, err := r.collectDescendantIDs(ctx, tx, parent.ID)
		if err != nil {
			return err
		}

		conds := make([]customfields.ShowIfCondition, 0, len(descendants)+1)
		conds = append(conds, customfields.ShowIfCondition{
			Field: "categoryId",
			Op:    "eq",
			Value: parent.ID,
		})
		for _, id := range descendants {
			conds = append(conds, customfields.ShowIfCondition{
				Field: "categoryId",
				Op:    "eq",
				Value: id,
			})
		}

		if _, err = r.upsertCollection(ctx, tx, parent, conds); err != nil {
			return err
		}

		parentID = parent.ParentID
	}

	return nil
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
