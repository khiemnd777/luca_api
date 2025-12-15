package collectionutils

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/metadata/customfields"
)

type TreeNode struct {
	ID           int
	ParentID     *int
	Name         *string
	CollectionID *int
}

type TreeConfig struct {
	TableName      string // categories, folders, ...
	IDColumn       string // id
	ParentIDColumn string // parent_id

	ShowIfFieldName  string // categoryId, folderId, ...
	CollectionGroup  string // category, folder
	CollectionPrefix string // category, folder
}

func UpsertAncestorCollections(
	ctx context.Context,
	tx *generated.Tx,
	cfg TreeConfig,
	nodeID int,
) error {

	node, err := fetchTreeNode(ctx, tx, cfg, nodeID)
	if err != nil {
		return err
	}
	if node.ParentID == nil {
		return nil
	}

	visited := map[int]bool{}
	curID := *node.ParentID

	for {
		if visited[curID] {
			return fmt.Errorf("cycle detected at node %d", curID)
		}
		visited[curID] = true

		parent, err := fetchTreeNode(ctx, tx, cfg, curID)
		if err != nil {
			return err
		}

		descendants, err := CollectDescendantIDs(ctx, tx, cfg, parent.ID)
		if err != nil {
			return err
		}

		conds := make([]customfields.ShowIfCondition, 0, len(descendants)+1)
		conds = append(conds, customfields.ShowIfCondition{
			Field: cfg.ShowIfFieldName,
			Op:    "eq",
			Value: parent.ID,
		})

		for _, id := range descendants {
			conds = append(conds, customfields.ShowIfCondition{
				Field: cfg.ShowIfFieldName,
				Op:    "eq",
				Value: id,
			})
		}

		if err := UpsertCollectionForNode(ctx, tx, cfg, parent, conds); err != nil {
			return err
		}

		if parent.ParentID == nil {
			break
		}
		curID = *parent.ParentID
	}

	return nil
}

func UpsertCollectionForNode(
	ctx context.Context,
	tx *generated.Tx,
	cfg TreeConfig,
	node *TreeNode,
	conds []customfields.ShowIfCondition,
) error {

	// ===== Fallback logic (BIT-EXACT) =====
	if len(conds) == 0 {

		// ROOT node → self + descendants (same as old Update)
		if node.ParentID == nil {
			descendants, err := CollectDescendantIDs(ctx, tx, cfg, node.ID)
			if err != nil {
				return err
			}

			conds = make([]customfields.ShowIfCondition, 0, len(descendants)+1)
			conds = append(conds, customfields.ShowIfCondition{
				Field: cfg.ShowIfFieldName,
				Op:    "eq",
				Value: node.ID,
			})

			for _, id := range descendants {
				conds = append(conds, customfields.ShowIfCondition{
					Field: cfg.ShowIfFieldName,
					Op:    "eq",
					Value: id,
				})
			}

		} else {
			// NON-root → self only (same as old Create + Update)
			conds = []customfields.ShowIfCondition{{
				Field: cfg.ShowIfFieldName,
				Op:    "eq",
				Value: node.ID,
			}}
		}
	}

	// ===== Persist collection =====

	showIf := customfields.ShowIfCondition{Any: conds}
	showIfJSON, err := json.Marshal(showIf)
	if err != nil {
		return err
	}

	slug := cfg.CollectionPrefix + "-" + strconv.Itoa(node.ID)

	name := cfg.CollectionPrefix
	if node.Name != nil && *node.Name != "" {
		name = *node.Name
	}

	// UPDATE existing collection
	if node.CollectionID != nil {
		_, err = tx.ExecContext(ctx, `
			UPDATE collections
			SET slug = $1,
			    name = $2,
			    show_if = $3,
			    integration = true,
			    "group" = $4
			WHERE id = $5
		`, slug, name, string(showIfJSON), cfg.CollectionGroup, *node.CollectionID)
		return err
	}

	// INSERT new collection
	rows, err := tx.QueryContext(ctx, `
		INSERT INTO collections (slug, name, show_if, integration, "group")
		VALUES ($1, $2, $3, true, $4)
		RETURNING id
	`, slug, name, string(showIfJSON), cfg.CollectionGroup)
	if err != nil {
		return err
	}
	defer rows.Close()

	if !rows.Next() {
		return fmt.Errorf("failed to insert collection for node %d", node.ID)
	}

	var collectionID int
	if err := rows.Scan(&collectionID); err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, fmt.Sprintf(`
		UPDATE %s
		SET collection_id = $1
		WHERE id = $2
	`, cfg.TableName), collectionID, node.ID)

	return err
}

func CollectDescendantIDs(
	ctx context.Context,
	tx *generated.Tx,
	cfg TreeConfig,
	parentID int,
) ([]int, error) {

	queue := []int{parentID}
	seen := map[int]struct{}{parentID: {}}
	var result []int

	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]

		rows, err := tx.QueryContext(ctx, fmt.Sprintf(`
			SELECT id
			FROM %s
			WHERE %s = $1
			  AND deleted_at IS NULL
		`, cfg.TableName, cfg.ParentIDColumn), id)
		if err != nil {
			return nil, err
		}

		for rows.Next() {
			var cid int
			if err := rows.Scan(&cid); err != nil {
				rows.Close()
				return nil, err
			}
			if _, ok := seen[cid]; ok {
				continue
			}
			seen[cid] = struct{}{}
			result = append(result, cid)
			queue = append(queue, cid)
		}
		rows.Close()
	}

	return result, nil
}

func fetchTreeNode(
	ctx context.Context,
	tx *generated.Tx,
	cfg TreeConfig,
	id int,
) (*TreeNode, error) {

	query := fmt.Sprintf(`
		SELECT id, parent_id, name, collection_id
		FROM %s
		WHERE id = $1 AND deleted_at IS NULL
	`, cfg.TableName)

	rows, err := tx.QueryContext(ctx, query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, sql.ErrNoRows
	}

	var node TreeNode
	var parentID sql.NullInt64
	var name sql.NullString
	var collectionID sql.NullInt64

	if err := rows.Scan(&node.ID, &parentID, &name, &collectionID); err != nil {
		return nil, err
	}

	if parentID.Valid {
		v := int(parentID.Int64)
		node.ParentID = &v
	}
	if name.Valid {
		node.Name = &name.String
	}
	if collectionID.Valid {
		v := int(collectionID.Int64)
		node.CollectionID = &v
	}

	return &node, nil
}
