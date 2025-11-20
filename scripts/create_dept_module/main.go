package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// ---------- Helpers ----------

func toPascal(s string) string {
	parts := strings.Split(s, "_")
	out := ""
	for _, p := range parts {
		if p == "" {
			continue
		}
		out += strings.ToUpper(p[:1]) + strings.ToLower(p[1:])
	}
	return out
}

// ---------- Templates ----------

// DTO: modules/main/features/__model/{module}_dto.go
// package model
func dtoTemplate(moduleSnake, structName string, ignoreCF bool) string {
	var cfLine string
	if !ignoreCF {
		cfLine = "\tCustomFields map[string]any `json:\"custom_fields,omitempty\"`\n"
	}

	return fmt.Sprintf(`package model

import "time"

type %sDTO struct {
	ID        int        `+"`json:\"id,omitempty\"`"+`
	Code      *string    `+"`json:\"code,omitempty\"`"+`
	Name      *string    `+"`json:\"name,omitempty\"`"+`
%s	CreatedAt time.Time `+"`json:\"created_at\"`"+`
	UpdatedAt time.Time `+"`json:\"updated_at\"`"+`
}
`, structName, cfLine)
}

// Repository: modules/main/features/{module}/repository/repository.go
// package repository
func repositoryTemplate(moduleSnake, structName string) string {
	template := `
package repository

import (
	"context"
	"time"

	"github.com/khiemnd777/andy_api/modules/main/config"
	model "github.com/khiemnd777/andy_api/modules/main/features/__model"
	relation "github.com/khiemnd777/andy_api/modules/main/features/__relation/policy"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/{{moduleSnake}}"
	dbutils "github.com/khiemnd777/andy_api/shared/db/utils"
	"github.com/khiemnd777/andy_api/shared/mapper"
	"github.com/khiemnd777/andy_api/shared/metadata/customfields"
	"github.com/khiemnd777/andy_api/shared/module"
	"github.com/khiemnd777/andy_api/shared/utils/table"
)

type {{Module}}Repository interface {
	Create(ctx context.Context, input model.{{Module}}DTO) (*model.{{Module}}DTO, error)
	Update(ctx context.Context, input model.{{Module}}DTO) (*model.{{Module}}DTO, error)
	GetByID(ctx context.Context, id int) (*model.{{Module}}DTO, error)
	List(ctx context.Context, query table.TableQuery) (table.TableListResult[model.{{Module}}DTO], error)
	Search(ctx context.Context, query dbutils.SearchQuery) (dbutils.SearchResult[model.{{Module}}DTO], error)
	Delete(ctx context.Context, id int) error
}

type {{moduleSnake}}Repo struct {
	db    *generated.Client
	deps  *module.ModuleDeps[config.ModuleConfig]
	cfMgr *customfields.Manager
}

func New{{Module}}Repository(db *generated.Client, deps *module.ModuleDeps[config.ModuleConfig], cfMgr *customfields.Manager) {{Module}}Repository {
	return &{{moduleSnake}}Repo{db: db, deps: deps, cfMgr: cfMgr}
}

func (r *{{moduleSnake}}Repo) Create(ctx context.Context, input model.{{Module}}DTO) (*model.{{Module}}DTO, error) {
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

	q := tx.{{Module}}.Create().
		SetNillableCode(input.Code).
		SetNillableName(input.Name)

	err = customfields.SetCustomFields(ctx, r.cfMgr, "{{moduleSnake}}", input.CustomFields, q, false)
	if err != nil {
		return nil, err
	}

	entity, err := q.Save(ctx)
	if err != nil {
		return nil, err
	}

	dto := mapper.MapAs[*generated.{{Module}}, *model.{{Module}}DTO](entity)
	
	_, err = relation.Upsert(ctx, tx, "{{moduleSnake}}", entity, input, dto)
	if err != nil {
		return nil, err
	}

	return dto, nil
}

func (r *{{moduleSnake}}Repo) Update(ctx context.Context, input model.{{Module}}DTO) (*model.{{Module}}DTO, error) {
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

	q := tx.{{Module}}.UpdateOneID(input.ID).
		SetNillableCode(input.Code).
		SetNillableName(input.Name)

	err = customfields.SetCustomFields(ctx, r.cfMgr, "{{moduleSnake}}", input.CustomFields, q, false)
	if err != nil {
		return nil, err
	}

	entity, err := q.Save(ctx)
	if err != nil {
		return nil, err
	}

	dto := mapper.MapAs[*generated.{{Module}}, *model.{{Module}}DTO](entity)
	
	_, err = relation.Upsert(ctx, tx, "{{moduleSnake}}", entity, input, dto)
	if err != nil {
		return nil, err
	}

	return dto, nil
}

func (r *{{moduleSnake}}Repo) GetByID(ctx context.Context, id int) (*model.{{Module}}DTO, error) {
	q := r.db.{{Module}}.Query().
		Where(
			{{moduleSnake}}.ID(id),
			{{moduleSnake}}.DeletedAtIsNil(),
		)

	entity, err := q.Only(ctx)
	if err != nil {
		return nil, err
	}

	dto := mapper.MapAs[*generated.{{Module}}, *model.{{Module}}DTO](entity)
	return dto, nil
}

func (r *{{moduleSnake}}Repo) List(ctx context.Context, query table.TableQuery) (table.TableListResult[model.{{Module}}DTO], error) {
	list, err := table.TableList(
		ctx,
		r.db.{{Module}}.Query().
			Where({{moduleSnake}}.DeletedAtIsNil()),
		query,
		{{moduleSnake}}.Table,
		{{moduleSnake}}.FieldID,
		{{moduleSnake}}.FieldID,
		func(src []*generated.{{Module}}) []*model.{{Module}}DTO {
			return mapper.MapListAs[*generated.{{Module}}, *model.{{Module}}DTO](src)
		},
	)
	if err != nil {
		var zero table.TableListResult[model.{{Module}}DTO]
		return zero, err
	}
	return list, nil
}

func (r *{{moduleSnake}}Repo) Search(ctx context.Context, query dbutils.SearchQuery) (dbutils.SearchResult[model.{{Module}}DTO], error) {
	return dbutils.Search(
		ctx,
		r.db.{{Module}}.Query().
			Where({{moduleSnake}}.DeletedAtIsNil()),
		[]string{
			dbutils.GetNormField({{moduleSnake}}.FieldCode),
			dbutils.GetNormField({{moduleSnake}}.FieldName),
		},
		query,
		{{moduleSnake}}.Table,
		{{moduleSnake}}.FieldID,
		{{moduleSnake}}.FieldID,
		{{moduleSnake}}.Or,
		func(src []*generated.{{Module}}) []*model.{{Module}}DTO {
			return mapper.MapListAs[*generated.{{Module}}, *model.{{Module}}DTO](src)
		},
	)
}

func (r *{{moduleSnake}}Repo) Delete(ctx context.Context, id int) error {
	return r.db.{{Module}}.UpdateOneID(id).
		SetDeletedAt(time.Now()).
		Exec(ctx)
}
`
	template = strings.ReplaceAll(template, "{{moduleSnake}}", moduleSnake)
	template = strings.ReplaceAll(template, "{{Module}}", structName)
	return template
}

// Service: modules/main/features/{module}/service/service.go
// package service
func serviceTemplate(moduleSnake, structName string) string {
	template := `
package service

import (
	"context"
	"fmt"

	"github.com/khiemnd777/andy_api/modules/main/config"
	model "github.com/khiemnd777/andy_api/modules/main/features/__model"
	"github.com/khiemnd777/andy_api/modules/main/features/{{moduleSnake}}/repository"
	"github.com/khiemnd777/andy_api/shared/cache"
	dbutils "github.com/khiemnd777/andy_api/shared/db/utils"
	"github.com/khiemnd777/andy_api/shared/metadata/customfields"
	"github.com/khiemnd777/andy_api/shared/module"
	searchmodel "github.com/khiemnd777/andy_api/shared/modules/search/model"
	"github.com/khiemnd777/andy_api/shared/pubsub"
	searchutils "github.com/khiemnd777/andy_api/shared/search"
	"github.com/khiemnd777/andy_api/shared/utils"
	"github.com/khiemnd777/andy_api/shared/utils/table"
)

type {{Module}}Service interface {
	Create(ctx context.Context, deptID int, input model.{{Module}}DTO) (*model.{{Module}}DTO, error)
	Update(ctx context.Context, deptID int, input model.{{Module}}DTO) (*model.{{Module}}DTO, error)
	GetByID(ctx context.Context, id int) (*model.{{Module}}DTO, error)
	List(ctx context.Context, query table.TableQuery) (table.TableListResult[model.{{Module}}DTO], error)
	Search(ctx context.Context, query dbutils.SearchQuery) (dbutils.SearchResult[model.{{Module}}DTO], error)
	Delete(ctx context.Context, id int) error
}

type {{moduleSnake}}Service struct {
	repo  repository.{{Module}}Repository
	deps  *module.ModuleDeps[config.ModuleConfig]
	cfMgr *customfields.Manager
}

func New{{Module}}Service(repo repository.{{Module}}Repository, deps *module.ModuleDeps[config.ModuleConfig], cfMgr *customfields.Manager) {{Module}}Service {
	return &{{moduleSnake}}Service{repo: repo, deps: deps, cfMgr: cfMgr}
}

// ----------------------------------------------------------------------------
// Cache Keys
// ----------------------------------------------------------------------------

func k{{Module}}ByID(id int) string {
	return fmt.Sprintf("{{moduleSnake}}:id:%d", id)
}

func k{{Module}}All() []string {
	return []string{
		k{{Module}}ListAll(),
		k{{Module}}SearchAll(),
	}
}

func k{{Module}}ListAll() string {
	return "{{moduleSnake}}:list:*"
}

func k{{Module}}SearchAll() string {
	return "{{moduleSnake}}:search:*"
}

func k{{Module}}List(q table.TableQuery) string {
	orderBy := ""
	if q.OrderBy != nil {
		orderBy = *q.OrderBy
	}
	return fmt.Sprintf("{{moduleSnake}}:list:l%d:p%d:o%s:d%s", q.Limit, q.Page, orderBy, q.Direction)
}

func k{{Module}}Search(q dbutils.SearchQuery) string {
	orderBy := ""
	if q.OrderBy != nil {
		orderBy = *q.OrderBy
	}
	return fmt.Sprintf("{{moduleSnake}}:search:k%s:l%d:p%d:o%s:d%s", q.Keyword, q.Limit, q.Page, orderBy, q.Direction)
}

// ----------------------------------------------------------------------------
// Create
// ----------------------------------------------------------------------------

func (s *{{moduleSnake}}Service) Create(ctx context.Context, deptID int, input model.{{Module}}DTO) (*model.{{Module}}DTO, error) {
	dto, err := s.repo.Create(ctx, input)
	if err != nil {
		return nil, err
	}

	if dto != nil && dto.ID > 0 {
		cache.InvalidateKeys(k{{Module}}ByID(dto.ID))
	}
	cache.InvalidateKeys(k{{Module}}All()...)

	s.upsertSearch(ctx, deptID, dto)

	return dto, nil
}

// ----------------------------------------------------------------------------
// Update
// ----------------------------------------------------------------------------

func (s *{{moduleSnake}}Service) Update(ctx context.Context, deptID int, input model.{{Module}}DTO) (*model.{{Module}}DTO, error) {
	dto, err := s.repo.Update(ctx, input)
	if err != nil {
		return nil, err
	}

	if dto != nil {
		cache.InvalidateKeys(k{{Module}}ByID(dto.ID))
	}
	cache.InvalidateKeys(k{{Module}}All()...)

	s.upsertSearch(ctx, deptID, dto)

	return dto, nil
}

// ----------------------------------------------------------------------------
// upsertSearch
// ----------------------------------------------------------------------------

func (s *{{moduleSnake}}Service) upsertSearch(ctx context.Context, deptID int, dto *model.{{Module}}DTO) {
	// Bạn có thể chỉnh lại cho phù hợp với module thực tế (Title/Content/Keywords...).
	kwPtr, _ := searchutils.BuildKeywords(ctx, s.cfMgr, "{{moduleSnake}}", []any{dto.Code}, dto.CustomFields)

	pubsub.PublishAsync("search:upsert", &searchmodel.Doc{
		EntityType: "{{moduleSnake}}",
		EntityID:   int64(dto.ID),
		Title:      *dto.Name,
		Subtitle:   nil,     
		Keywords:   &kwPtr,
		Content:    nil,     
		Attributes: map[string]any{},
		OrgID:      utils.Ptr(int64(deptID)),
		OwnerID:    nil,
	})
}

// ----------------------------------------------------------------------------
// GetByID
// ----------------------------------------------------------------------------

func (s *{{moduleSnake}}Service) GetByID(ctx context.Context, id int) (*model.{{Module}}DTO, error) {
	return cache.Get(k{{Module}}ByID(id), cache.TTLMedium, func() (*model.{{Module}}DTO, error) {
		return s.repo.GetByID(ctx, id)
	})
}

// ----------------------------------------------------------------------------
// List
// ----------------------------------------------------------------------------

func (s *{{moduleSnake}}Service) List(ctx context.Context, q table.TableQuery) (table.TableListResult[model.{{Module}}DTO], error) {
	type boxed = table.TableListResult[model.{{Module}}DTO]
	key := k{{Module}}List(q)

	ptr, err := cache.Get(key, cache.TTLMedium, func() (*boxed, error) {
		res, e := s.repo.List(ctx, q)
		if e != nil {
			return nil, e
		}
		return &res, nil
	})
	if err != nil {
		var zero boxed
		return zero, err
	}
	return *ptr, nil
}

// ----------------------------------------------------------------------------
// Delete
// ----------------------------------------------------------------------------

func (s *{{moduleSnake}}Service) Delete(ctx context.Context, id int) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	cache.InvalidateKeys(k{{Module}}All()...)
	cache.InvalidateKeys(k{{Module}}ByID(id))
	return nil
}

// ----------------------------------------------------------------------------
// Search
// ----------------------------------------------------------------------------

func (s *{{moduleSnake}}Service) Search(ctx context.Context, q dbutils.SearchQuery) (dbutils.SearchResult[model.{{Module}}DTO], error) {
	type boxed = dbutils.SearchResult[model.{{Module}}DTO]
	key := k{{Module}}Search(q)

	ptr, err := cache.Get(key, cache.TTLMedium, func() (*boxed, error) {
		res, e := s.repo.Search(ctx, q)
		if e != nil {
			return nil, e
		}
		return &res, nil
	})
	if err != nil {
		var zero boxed
		return zero, err
	}
	return *ptr, nil
}
`
	template = strings.ReplaceAll(template, "{{moduleSnake}}", moduleSnake)
	template = strings.ReplaceAll(template, "{{Module}}", structName)
	return template
}

// Handler: modules/main/features/{module}/handler/handler.go
// package handler
func handlerTemplate(moduleSnake, structName string) string {
	// moduleSnake: clinic, product, ...
	// structName : Clinic, Product, ...

	return fmt.Sprintf(`package handler

import (
	"github.com/gofiber/fiber/v2"

	"github.com/khiemnd777/andy_api/modules/main/config"
	model "github.com/khiemnd777/andy_api/modules/main/features/__model"
	"github.com/khiemnd777/andy_api/modules/main/features/%[1]s/service"
	"github.com/khiemnd777/andy_api/shared/app"
	"github.com/khiemnd777/andy_api/shared/app/client_error"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	dbutils "github.com/khiemnd777/andy_api/shared/db/utils"
	"github.com/khiemnd777/andy_api/shared/middleware/rbac"
	"github.com/khiemnd777/andy_api/shared/module"
	"github.com/khiemnd777/andy_api/shared/utils"
	"github.com/khiemnd777/andy_api/shared/utils/table"
)

type %[2]sHandler struct {
	svc  service.%[2]sService
	deps *module.ModuleDeps[config.ModuleConfig]
}

func New%[2]sHandler(svc service.%[2]sService, deps *module.ModuleDeps[config.ModuleConfig]) *%[2]sHandler {
	return &%[2]sHandler{svc: svc, deps: deps}
}

func (h *%[2]sHandler) RegisterRoutes(router fiber.Router) {
	app.RouterGet(router, "/:dept_id<int>/%[1]s/list", h.List)
	app.RouterGet(router, "/:dept_id<int>/%[1]s/search", h.Search)
	app.RouterGet(router, "/:dept_id<int>/%[1]s/:id<int>", h.GetByID)
	app.RouterPost(router, "/:dept_id<int>/%[1]s", h.Create)
	app.RouterPut(router, "/:dept_id<int>/%[1]s/:id<int>", h.Update)
	app.RouterDelete(router, "/:dept_id<int>/%[1]s/:id<int>", h.Delete)
}

func (h *%[2]sHandler) List(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), "%[1]s.view"); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}
	q := table.ParseTableQuery(c, 20)
	res, err := h.svc.List(c.UserContext(), q)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.Status(fiber.StatusOK).JSON(res)
}

func (h *%[2]sHandler) Search(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), "%[1]s.view"); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}
	q := dbutils.ParseSearchQuery(c, 20)
	res, err := h.svc.Search(c.UserContext(), q)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.Status(fiber.StatusOK).JSON(res)
}

func (h *%[2]sHandler) GetByID(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), "%[1]s.view"); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}
	id, _ := utils.GetParamAsInt(c, "id")
	if id <= 0 {
		return client_error.ResponseError(c, fiber.StatusBadRequest, nil, "invalid id")
	}

	dto, err := h.svc.GetByID(c.UserContext(), id)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.Status(fiber.StatusOK).JSON(dto)
}

func (h *%[2]sHandler) Create(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), "%[1]s.create"); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}
	var payload model.%[2]sDTO
	if err := c.BodyParser(&payload); err != nil {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, "invalid body")
 	}
	// phần validate tuỳ module, tạm giữ nguyên generic.

	deptID, _ := utils.GetDeptIDInt(c)

	dto, err := h.svc.Create(c.UserContext(), deptID, payload)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.Status(fiber.StatusCreated).JSON(dto)
}

func (h *%[2]sHandler) Update(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), "%[1]s.update"); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}
	id, _ := utils.GetParamAsInt(c, "id")
	if id <= 0 {
		return client_error.ResponseError(c, fiber.StatusBadRequest, nil, "invalid id")
	}

	var payload model.%[2]sDTO
	if err := c.BodyParser(&payload); err != nil {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, "invalid body")
	}
	payload.ID = id

	deptID, _ := utils.GetDeptIDInt(c)

	dto, err := h.svc.Update(c.UserContext(), deptID, payload)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.Status(fiber.StatusOK).JSON(dto)
}

func (h *%[2]sHandler) Delete(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), "%[1]s.delete"); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}
	id, _ := utils.GetParamAsInt(c, "id")
	if id <= 0 {
		return client_error.ResponseError(c, fiber.StatusBadRequest, nil, "invalid id")
	}
	if err := h.svc.Delete(c.UserContext(), id); err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.SendStatus(fiber.StatusNoContent)
}
`, moduleSnake, structName)
}

// Registry: modules/main/features/{module}/registry.go
// package {module}
func registryTemplate(moduleSnake, structName string) string {
	return fmt.Sprintf(`package %s

import (
	"github.com/gofiber/fiber/v2"

	"github.com/khiemnd777/andy_api/modules/main/config"
	"github.com/khiemnd777/andy_api/modules/main/features/%[1]s/handler"
	"github.com/khiemnd777/andy_api/modules/main/features/%[1]s/repository"
	"github.com/khiemnd777/andy_api/modules/main/features/%[1]s/service"
	"github.com/khiemnd777/andy_api/modules/main/registry"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/metadata/customfields"
	"github.com/khiemnd777/andy_api/shared/module"
)

type feature struct{}

func (feature) ID() string    { return "%[1]s" }
func (feature) Priority() int { return 60 }

func (feature) Register(router fiber.Router, deps *module.ModuleDeps[config.ModuleConfig], cfMgr *customfields.Manager) error {
	repo := repository.New%[2]sRepository(deps.Ent.(*generated.Client), deps, cfMgr)
	svc := service.New%[2]sService(repo, deps, cfMgr)
	h := handler.New%[2]sHandler(svc, deps)
	h.RegisterRoutes(router)
	return nil
}

func init() { registry.Register(feature{}) }
`, moduleSnake, structName)
}

// ---------- Ent Schema Template ----------

// shared/db/ent/schema/dept_{module}.go
func entSchemaTemplate(moduleSnake, structName string) string {
	return fmt.Sprintf(`package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type %s struct {
	ent.Schema
}

func (%s) Fields() []ent.Field {
	return []ent.Field{
		field.String("code").
			Optional().
			Nillable(),

		field.String("name").
			Optional().
			Nillable(),

		field.Bool("active").
			Default(true),

		field.JSON("custom_fields", map[string]any{}).
			Optional().
			Default(map[string]any{}),

		field.Time("created_at").
			Default(time.Now).
			Immutable(),

		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),

		field.Time("deleted_at").
			Optional().
			Nillable(),
	}
}

func (%s) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("id", "deleted_at"),
		index.Fields("code"),
		index.Fields("code", "deleted_at").Unique(),
		index.Fields("name", "deleted_at"),
		index.Fields("deleted_at"),
	}
}
`, structName, structName, structName)
}

// ---------- Flyway Templates ----------

func flywayIndexesTemplate(moduleSnake string) string {
	table := moduleSnake + "s"
	return fmt.Sprintf(`CREATE INDEX IF NOT EXISTS ix_%[1]s_id_not_deleted
  ON %[2]s(id)
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS ix_%[1]s_code_not_deleted
  ON %[2]s(code)
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS ix_%[1]s_name_not_deleted
  ON %[2]s(name)
  WHERE deleted_at IS NULL;
`, moduleSnake, table)
}

func flywayNormFieldsTemplate(moduleSnake string) string {
	table := moduleSnake + "s"
	return fmt.Sprintf(`CREATE EXTENSION IF NOT EXISTS unaccent;
CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE OR REPLACE FUNCTION public.unaccent_immutable(text)
RETURNS text
LANGUAGE sql IMMUTABLE PARALLEL SAFE RETURNS NULL ON NULL INPUT
AS $$ SELECT unaccent('unaccent'::regdictionary, $1) $$;

ALTER TABLE %[2]s
  ADD COLUMN IF NOT EXISTS code_norm text GENERATED ALWAYS AS (unaccent_immutable(lower(code))) STORED,
	ADD COLUMN IF NOT EXISTS name_norm text GENERATED ALWAYS AS (unaccent_immutable(lower(name))) STORED;

CREATE INDEX IF NOT EXISTS idx_%[1]s_code_trgm_norm  ON %[2]s USING gin (code_norm gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_%[1]s_name_trgm_norm  ON %[2]s USING gin (name_norm gin_trgm_ops);
`, moduleSnake, table)
}

// NEW: custom_fields migration
func flywayCustomFieldsTemplate(moduleSnake string) string {
	table := moduleSnake + "s"
	return fmt.Sprintf(`ALTER TABLE %s
  ADD COLUMN IF NOT EXISTS custom_fields JSONB DEFAULT '{}'::jsonb;

CREATE INDEX IF NOT EXISTS idx_%s_custom_fields_gin ON %s USING GIN (custom_fields);
`, table, table, table)
}

// NEW: RBAC matrix migration
func flywayRBACMatrixTemplate(moduleSnake, labelFlag string) string {
	return fmt.Sprintf(`-- ============================================
-- RBAC PERMISSIONS + ADMIN ROLE UPSERT SCRIPT
-- ============================================

-- 1. Ensure role "admin" exists
INSERT INTO roles (role_name)
VALUES ('admin')
ON CONFLICT (role_name)
DO UPDATE SET role_name = EXCLUDED.role_name;

-- ============================================
-- PERMISSIONS UPSERT
-- ============================================
INSERT INTO permissions (permission_name, permission_value)
VALUES
  ('%[2]s - Xem', '%[1]s.view'),
  ('%[2]s - Tạo', '%[1]s.create'),
  ('%[2]s - Sửa', '%[1]s.update'),
  ('%[2]s - Xoá', '%[1]s.delete'),
  ('%[2]s - Tìm kiếm', '%[1]s.search'),
	('%[2]s - Import', '%[1]s.import'),
	('%[2]s - Export', '%[1]s.export')
ON CONFLICT (permission_value)
DO UPDATE SET permission_name = EXCLUDED.permission_name;

-- ============================================
-- LINK ALL PERMISSIONS TO ADMIN ROLE
-- ============================================
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.permission_value IN (
  '%[1]s.view',
  '%[1]s.create',
  '%[1]s.update',
  '%[1]s.delete',
  '%[1]s.search',
	'%[1]s.import',
	'%[1]s.export'
)
WHERE r.role_name = 'admin'
ON CONFLICT DO NOTHING;
`, moduleSnake, labelFlag)
}

func flywayCollectionsTemplate(moduleSlug, label string) string {
	return fmt.Sprintf(`INSERT INTO collections (slug, name)
VALUES ('%s', '%s')
ON CONFLICT (slug)
DO UPDATE SET name = EXCLUDED.name;
`, moduleSlug, label)
}

func getLastFlywayVersion(dir string) (int, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0, err
	}
	maxVer := 0
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasPrefix(name, "V") || !strings.HasSuffix(name, ".sql") {
			continue
		}
		base := strings.TrimSuffix(name, ".sql")
		parts := strings.SplitN(base, "__", 2)
		if len(parts) == 0 {
			continue
		}
		numStr := strings.TrimPrefix(parts[0], "V")
		n, err := strconv.Atoi(numStr)
		if err != nil {
			continue
		}
		if n > maxVer {
			maxVer = n
		}
	}
	return maxVer, nil
}

// ---------- MAIN ----------

func main() {
	moduleFlag := flag.String("module", "", "Module name, e.g., clinic, dentist,...")
	ignoreCF := flag.Bool("ignorecf", false, "Ignore CustomFields in DTO")
	labelFlag := flag.String("label", "", "Label, e.g., Clinic, Dentist,...")
	flag.Parse()

	if *moduleFlag == "" {
		panic("Missing --module=name (e.g. --module=clinic)")
	}

	moduleSnake := strings.ToLower(*moduleFlag)
	structName := toPascal(moduleSnake)

	// label dùng cho collections.name (nếu không truyền thì fallback = structName)
	label := strings.TrimSpace(*labelFlag)
	if label == "" {
		label = structName
	}

	// Paths
	baseDir := filepath.Join("modules", "main", "features", moduleSnake)
	repoDir := filepath.Join(baseDir, "repository")
	svcDir := filepath.Join(baseDir, "service")
	handlerDir := filepath.Join(baseDir, "handler")
	modelDir := filepath.Join("modules", "main", "features", "__model")
	entSchemaDir := filepath.Join("shared", "db", "ent", "schema")
	flywayDir := filepath.Join("flyway", "sql")

	// Create dirs
	for _, dir := range []string{baseDir, repoDir, svcDir, handlerDir, modelDir, entSchemaDir, flywayDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			panic(err)
		}
	}

	write := func(path, content string) {
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			panic(err)
		}
		fmt.Println("Generated:", path)
	}

	// DTO
	dtoPath := filepath.Join(modelDir, moduleSnake+"_dto.go")
	write(dtoPath, dtoTemplate(moduleSnake, structName, *ignoreCF))

	// Repository
	repoPath := filepath.Join(repoDir, "repository.go")
	write(repoPath, repositoryTemplate(moduleSnake, structName))

	// Service
	svcPath := filepath.Join(svcDir, "service.go")
	write(svcPath, serviceTemplate(moduleSnake, structName))

	// Handler
	handlerPath := filepath.Join(handlerDir, "handler.go")
	write(handlerPath, handlerTemplate(moduleSnake, structName))

	// Registry
	registryPath := filepath.Join(baseDir, "registry.go")
	write(registryPath, registryTemplate(moduleSnake, structName))

	// Ent Schema: shared/db/ent/schema/dept_{module}.go
	entPath := filepath.Join(entSchemaDir, "dept_"+moduleSnake+".go")
	write(entPath, entSchemaTemplate(moduleSnake, structName))

	// Flyway migrations: đọc version cuối cùng và tạo 4 + 1 file mới
	lastVer, err := getLastFlywayVersion(flywayDir)
	if err != nil {
		panic(err)
	}

	// 1) Indexes
	nextVer := lastVer + 1
	indexesPath := filepath.Join(flywayDir, fmt.Sprintf("V%d__dept_%s_indexes.sql", nextVer, moduleSnake))
	write(indexesPath, flywayIndexesTemplate(moduleSnake))

	// 2) Norm fields
	normVer := nextVer + 1
	normPath := filepath.Join(flywayDir, fmt.Sprintf("V%d__dept_%s_norm_fields.sql", normVer, moduleSnake))
	write(normPath, flywayNormFieldsTemplate(moduleSnake))

	// 3) Custom fields
	cfVer := normVer + 1
	cfPath := filepath.Join(flywayDir, fmt.Sprintf("V%d__dept_%s_custom_fields.sql", cfVer, moduleSnake))
	write(cfPath, flywayCustomFieldsTemplate(moduleSnake))

	// 4) RBAC matrix
	rbacVer := cfVer + 1
	rbacPath := filepath.Join(flywayDir, fmt.Sprintf("V%d__dept_%s_rbac_matrix.sql", rbacVer, moduleSnake))
	write(rbacPath, flywayRBACMatrixTemplate(moduleSnake, label))

	// 5) Metadata collections
	collectionsVer := rbacVer + 1
	collectionsPath := filepath.Join(flywayDir, fmt.Sprintf("V%d__dept_%s_metadata_collections.sql", collectionsVer, moduleSnake))
	write(collectionsPath, flywayCollectionsTemplate(moduleSnake, label))

	fmt.Println("✔ Done at", time.Now())
}
