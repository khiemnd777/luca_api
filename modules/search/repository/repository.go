package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"slices"
	"strings"

	"github.com/khiemnd777/andy_api/modules/search/config"
	"github.com/khiemnd777/andy_api/modules/search/model"
	dbutils "github.com/khiemnd777/andy_api/shared/db/utils"
	"github.com/khiemnd777/andy_api/shared/module"
	sharedmodel "github.com/khiemnd777/andy_api/shared/modules/search/model"
)

type SearchRepository interface {
	Upsert(ctx context.Context, d sharedmodel.Doc) error
	Delete(ctx context.Context, entityType string, entityID int64) error
	Search(ctx context.Context, opt model.Options) ([]model.Row, error)
}

type searchRepo struct {
	deps *module.ModuleDeps[config.ModuleConfig]
}

func NewSearchRepository(deps *module.ModuleDeps[config.ModuleConfig]) SearchRepository {
	return &searchRepo{deps: deps}
}

func (r *searchRepo) Upsert(ctx context.Context, d sharedmodel.Doc) error {
	attr := []byte("{}")
	if d.Attributes != nil {
		b, _ := json.Marshal(d.Attributes)
		attr = b
	}
	q := `
INSERT INTO search_index
(entity_type, entity_id, title, subtitle, keywords, content, attributes, org_id, owner_id, acl_hash, updated_at)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10, now())
ON CONFLICT (entity_type, entity_id) DO UPDATE
SET title=$3, subtitle=$4, keywords=$5, content=$6, attributes=$7,
    org_id=$8, owner_id=$9, acl_hash=$10, updated_at=now();`
	_, err := r.deps.DB.ExecContext(ctx, q,
		d.EntityType, d.EntityID, d.Title, d.Subtitle, d.Keywords, d.Content, attr, d.OrgID, d.OwnerID, d.ACLHash,
	)
	if err != nil {
		log.Printf("[search] Upsert failed: %v", err)
	}
	return err
}

func (r *searchRepo) Delete(ctx context.Context, entityType string, entityID int64) error {
	_, err := r.deps.DB.ExecContext(ctx,
		`DELETE FROM search_index WHERE entity_type=$1 AND entity_id=$2`,
		entityType, entityID,
	)
	if err != nil {
		log.Printf("[search] Delete failed: %v", err)
	}
	return err
}

// Build attribute filters (simple exact match on root keys)
func buildAttrPreds(filters map[string]string, args *[]any) string {
	if len(filters) == 0 {
		return ""
	}
	parts := make([]string, 0, len(filters))
	for k, v := range filters {
		*args = append(*args, k, v)
		parts = append(parts, fmt.Sprintf(`attributes ->> $%d = $%d`, len(*args)-1, len(*args)))
	}
	return " AND (" + strings.Join(parts, " AND ") + ")"
}

// Search with full-text first; optional trigram fallback
func (r *searchRepo) Search(ctx context.Context, opt model.Options) ([]model.Row, error) {
	limit := opt.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	offset := opt.Offset
	args := make([]any, 0, 16)

	// Dynamic filters
	where := "TRUE"
	// Types
	if len(opt.Types) > 0 {
		where += " AND entity_type = ANY($1)"
		args = append(args, dbutils.PqStringArray(opt.Types)) // helper below
	}
	// Org/Owner
	if opt.OrgID != nil {
		args = append(args, *opt.OrgID)
		where += fmt.Sprintf(" AND (org_id IS NULL OR org_id = $%d)", len(args))
	}
	if opt.OwnerID != nil {
		args = append(args, *opt.OwnerID)
		where += fmt.Sprintf(" AND (owner_id IS NULL OR owner_id = $%d)", len(args))
	}
	// Attributes
	where += buildAttrPreds(opt.Filters, &args)

	// Full-text CTE
	args = append(args, opt.Query)     // $N-1
	args = append(args, limit, offset) // $N, $N+1
	ft := fmt.Sprintf(`
WITH q AS (SELECT plainto_tsquery('simple', unaccent($%d)) AS term)
SELECT entity_type, entity_id, title, subtitle, keywords, attributes, updated_at,
       ts_rank(tsv, (SELECT term FROM q)) AS rank,
       'ts' AS src
FROM search_index
WHERE %s
  AND tsv @@ (SELECT term FROM q)
ORDER BY rank DESC, updated_at DESC
LIMIT $%d OFFSET $%d`, len(args)-2, where, len(args)-1, len(args))

	// If not using fallback → run only full-text
	if !opt.UseTrgmFallback {
		return r.scanRows(ctx, ft, args...)
	}

	// Trigram fallback CTE
	args2 := slices.Clone(args)
	args2[len(args2)-2] = opt.Query // keep same $ positions
	tr := fmt.Sprintf(`
WITH q AS (SELECT unaccent(lower($%d)) AS uq)
SELECT entity_type, entity_id, title, subtitle, keywords, attributes, updated_at,
       NULL::float AS rank,
       'trgm' AS src
FROM search_index
WHERE %s
  AND norm ILIKE '%%' || (SELECT uq FROM q) || '%%'
ORDER BY updated_at DESC
LIMIT $%d OFFSET $%d`, len(args2)-2, where, len(args2)-1, len(args2))

	// Union + distinct on (entity_type, entity_id), prefer ts
	union := fmt.Sprintf(`
WITH ts AS (%s),
     tg AS (%s)
SELECT DISTINCT ON (entity_type, entity_id)
  entity_type, entity_id, title, subtitle, keywords, attributes, updated_at, rank
FROM (
  SELECT * FROM ts
  UNION ALL
  SELECT * FROM tg
) u
ORDER BY entity_type, entity_id,
         CASE WHEN rank IS NULL THEN 1 ELSE 0 END,
         rank DESC NULLS LAST, updated_at DESC
LIMIT $%d OFFSET $%d`, ft, tr, len(args)-1, len(args)) // same limit/offset usage

	return r.scanRows(ctx, union, args...)
}

func (r *searchRepo) scanRows(ctx context.Context, q string, args ...any) ([]model.Row, error) {
	rows, err := r.deps.DB.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]model.Row, 0, 32)
	for rows.Next() {
		var (
			row     model.Row
			attrRaw []byte
		)
		if err := rows.Scan(&row.EntityType, &row.EntityID, &row.Title, &row.Subtitle, &row.Keywords, &attrRaw, &row.UpdatedAt, &row.Rank); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(attrRaw, &row.Attributes)
		out = append(out, row)
	}
	return out, rows.Err()
}
