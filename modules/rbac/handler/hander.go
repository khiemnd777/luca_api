package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/khiemnd777/andy_api/modules/rbac/service"
	"github.com/khiemnd777/andy_api/shared/app"
	"github.com/khiemnd777/andy_api/shared/app/client_error"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	dbutils "github.com/khiemnd777/andy_api/shared/db/utils"
	"github.com/khiemnd777/andy_api/shared/logger"
	"github.com/khiemnd777/andy_api/shared/middleware/rbac"
	"github.com/khiemnd777/andy_api/shared/utils"
	"github.com/khiemnd777/andy_api/shared/utils/table"
)

type RBACHandler struct {
	db  *generated.Client
	svc service.RBACService
}

func NewRBACHandler(db *generated.Client, svc service.RBACService) *RBACHandler {
	return &RBACHandler{db: db, svc: svc}
}

// ---------- DTOs
type createOrUpdateRoleReq struct {
	ID          int    `json:"id"`
	RoleName    string `json:"role_name"`
	DisplayName string `json:"display_name"`
	Brief       string `json:"brief"`
}
type renameRoleReq struct {
	RoleName string `json:"role_name"`
}

type createPermReq struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}
type matrixReq struct {
	RoleID  int   `json:"role_id"`
	PermIDs []int `json:"perm_ids"`
}
type matrixDeltaReq struct {
	RoleID  int   `json:"role_id"`
	PermIDs []int `json:"perm_ids"`
}

func (h *RBACHandler) RegisterRoutes(router fiber.Router) {
	// Role
	app.RouterGet(router, "/roles/me", h.ListRolesOfMe)
	app.RouterPost(router, "/roles", h.CreateRole)
	app.RouterGet(router, "/roles", h.ListRoles)
	app.RouterGet(router, "/user/:user_id<int>/roles", h.ListRolesByUserID)
	app.RouterGet(router, "/roles/search", h.SearchRoles)
	app.RouterPut(router, "/roles/:id/rename", h.RenameRole)
	app.RouterPut(router, "/roles/:id", h.UpdateRole)
	app.RouterGet(router, "/roles/:id", h.GetRole)
	app.RouterDelete(router, "/roles/:id", h.DeleteRole)

	// Permission
	app.RouterPost(router, "/permissions", h.CreatePermission)
	app.RouterGet(router, "/permissions", h.ListPermissions)
	app.RouterDelete(router, "/permissions/:id", h.DeletePermission)

	// Matrix
	app.RouterPost(router, "/matrix/replace", h.ReplaceRolePermissions)
	app.RouterPost(router, "/matrix/add", h.AddRolePermissions)
	app.RouterPost(router, "/matrix/remove", h.RemoveRolePermissions)
	app.RouterGet(router, "/matrix/me", h.GetMyPermissionMatrix)
	app.RouterGet(router, "/matrix/users/:id", h.GetUserPermissionMatrix)
	app.RouterGet(router, "/matrix/roles/:id", h.GetRolePermissions)
	app.RouterGet(router, "/matrix", h.GetMatrix)
}

// ---------- Handlers (all require rbac.manage)
func (h *RBACHandler) CreateRole(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.db, "rbac.manage"); err != nil {
		return c.Status(http.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
	}
	var req createOrUpdateRoleReq
	if err := c.BodyParser(&req); err != nil || req.RoleName == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid payload"})
	}
	id, err := h.svc.CreateRole(c.UserContext(), req.RoleName, req.DisplayName, req.Brief)
	if err != nil {
		logger.Error("CreateRole failed: " + err.Error())
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"id": id})
}

func (h *RBACHandler) RenameRole(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.db, "rbac.manage"); err != nil {
		return c.Status(http.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
	}
	var req renameRoleReq
	if err := c.BodyParser(&req); err != nil || req.RoleName == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid payload"})
	}
	roleID, err := c.ParamsInt("id")
	if err != nil || roleID <= 0 {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid role id"})
	}
	if err := h.svc.RenameRole(c.UserContext(), roleID, req.RoleName); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.SendStatus(http.StatusNoContent)
}

func (h *RBACHandler) UpdateRole(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.db, "rbac.manage"); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}
	var req createOrUpdateRoleReq
	if err := c.BodyParser(&req); err != nil || req.RoleName == "" {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, "Invalid payload")
	}
	roleID, err := c.ParamsInt("id")
	if err != nil || roleID <= 0 {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, "Invalid role id")
	}
	if err := h.svc.UpdateRole(c.UserContext(), roleID, req.RoleName, req.DisplayName, req.Brief); err != nil {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, err.Error())
	}
	return c.SendStatus(http.StatusNoContent)
}

func (h *RBACHandler) GetRole(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.db, "rbac.manage"); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}
	roleID, err := utils.GetParamAsInt(c, "id")
	if err != nil || roleID <= 0 {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, "invalid role id")
	}
	result, err := h.svc.GetRole(c.UserContext(), roleID)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.JSON(result)
}

func (h *RBACHandler) DeleteRole(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.db, "rbac.manage"); err != nil {
		return c.Status(http.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
	}
	roleID, err := c.ParamsInt("id")
	if err != nil || roleID <= 0 {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid role id"})
	}
	if err := h.svc.DeleteRole(c.UserContext(), roleID); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.SendStatus(http.StatusNoContent)
}

func (h *RBACHandler) ListRoles(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.db, "rbac.manage"); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}
	tablequery := table.ParseTableQuery(c, 50)
	data, err := h.svc.ListRoles(c.UserContext(), tablequery)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, err.Error())
	}
	return c.Status(fiber.StatusOK).JSON(data)
}

func (h *RBACHandler) ListRolesByUserID(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.db, "rbac.manage"); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}
	userID, _ := utils.GetParamAsInt(c, "user_id")
	tablequery := table.ParseTableQuery(c, 50)
	data, err := h.svc.ListRolesByUserID(c.UserContext(), userID, tablequery)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, err.Error())
	}
	return c.Status(fiber.StatusOK).JSON(data)
}

func (h *RBACHandler) SearchRoles(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.db, "rbac.manage"); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}

	searchquery := dbutils.ParseSearchQuery(c, 20)
	data, err := h.svc.SearchRoles(c.UserContext(), searchquery)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, err.Error())
	}
	return c.Status(fiber.StatusOK).JSON(data)
}

func (h *RBACHandler) ListRolesOfMe(c *fiber.Ctx) error {
	limit, offset := parsePaging(c, 50, 0)
	userID, _ := utils.GetUserIDInt(c)
	data, err := h.svc.ListUserRoles(c.UserContext(), userID, limit, offset)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(data)
}

func (h *RBACHandler) CreatePermission(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.db, "rbac.manage"); err != nil {
		return c.Status(http.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
	}
	var req createPermReq
	if err := c.BodyParser(&req); err != nil || req.Name == "" || req.Value == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid payload"})
	}
	id, err := h.svc.CreatePermission(c.UserContext(), req.Name, req.Value)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"id": id})
}

func (h *RBACHandler) DeletePermission(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.db, "rbac.manage"); err != nil {
		return c.Status(http.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
	}
	id, err := c.ParamsInt("id")
	if err != nil || id <= 0 {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid permission id"})
	}
	if err := h.svc.DeletePermission(c.UserContext(), id); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.SendStatus(http.StatusNoContent)
}

func (h *RBACHandler) ListPermissions(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.db, "rbac.manage"); err != nil {
		return c.Status(http.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
	}
	limit, offset := parsePaging(c, 100, 0)
	data, err := h.svc.ListPermissions(c.UserContext(), limit, offset)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(data)
}

// ---- Matrix
func (h *RBACHandler) ReplaceRolePermissions(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.db, "rbac.manage"); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}
	var req matrixReq
	if err := c.BodyParser(&req); err != nil || req.RoleID <= 0 {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, "invalid payload")
	}
	if err := h.svc.ReplaceRolePermissions(c.UserContext(), req.RoleID, req.PermIDs); err != nil {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, err.Error())
	}
	return c.SendStatus(http.StatusNoContent)
}
func (h *RBACHandler) AddRolePermissions(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.db, "rbac.manage"); err != nil {
		return c.Status(http.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
	}
	var req matrixDeltaReq
	if err := c.BodyParser(&req); err != nil || req.RoleID <= 0 || len(req.PermIDs) == 0 {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid payload"})
	}
	if err := h.svc.AddRolePermissions(c.UserContext(), req.RoleID, req.PermIDs); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.SendStatus(http.StatusNoContent)
}
func (h *RBACHandler) RemoveRolePermissions(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.db, "rbac.manage"); err != nil {
		return c.Status(http.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
	}
	var req matrixDeltaReq
	if err := c.BodyParser(&req); err != nil || req.RoleID <= 0 || len(req.PermIDs) == 0 {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid payload"})
	}
	if err := h.svc.RemoveRolePermissions(c.UserContext(), req.RoleID, req.PermIDs); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.SendStatus(http.StatusNoContent)
}
func (h *RBACHandler) GetRolePermissions(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.db, "rbac.manage"); err != nil {
		return c.Status(http.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
	}
	roleID, err := c.ParamsInt("id")
	if err != nil || roleID <= 0 {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid role id"})
	}

	ids, err := h.svc.GetRolePermissionIDs(c.UserContext(), roleID)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{
		"role_id":  roleID,
		"perm_ids": ids,
	})
}
func (h *RBACHandler) GetRolePermissionMatrix(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.db, "rbac.manage"); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}
	raw := utils.GetQueryAsString(c, "role_ids")
	if strings.TrimSpace(raw) == "" {
		return client_error.ResponseError(c, fiber.StatusBadRequest, nil, "role_ids is required (comma-separated)")
	}

	parts := strings.Split(raw, ",")
	roleIDs := make([]int, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		id, err := strconv.Atoi(p)
		if err != nil || id <= 0 {
			return client_error.ResponseError(c, fiber.StatusBadRequest, err, err.Error())
		}
		roleIDs = append(roleIDs, id)
	}
	if len(roleIDs) == 0 {
		return client_error.ResponseError(c, fiber.StatusBadRequest, nil, "No valid role_ids")
	}

	matrix, err := h.svc.GetRolePermissionMatrix(c.UserContext(), roleIDs)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, err.Error())
	}
	// matrix: map[int][]int. Đổi key sang string để JSON friendly nếu thích.
	out := fiber.Map{}
	for k, v := range matrix {
		out[strconv.Itoa(k)] = v
	}
	return c.JSON(out)
}

// ---- User Permission Matrix
//
// Example response:
//
//	{
//	  "permissions": [
//	    { "id": 1, "name": "User Read", "value": "user_read" },
//	    { "id": 2, "name": "User Write", "value": "user_write" },
//	    { "id": 3, "name": "Product Manage", "value": "product_manage" }
//	  ],
//	  "roles": [
//	    {
//	      "role_id": 10,
//	      "role_name": "admin",
//	      "flags": [true, true, true]
//	    },
//	    {
//	      "role_id": 12,
//	      "role_name": "editor",
//	      "flags": [true, false, true]
//	    }
//	  ]
//	}
func (h *RBACHandler) GetMyPermissionMatrix(c *fiber.Ctx) error {
	userID, ok := utils.GetUserIDInt(c)
	if !ok || userID <= 0 {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}
	matrix, err := h.svc.GetUserRolePermissionMatrix(c.UserContext(), userID)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(matrix)
}
func (h *RBACHandler) GetUserPermissionMatrix(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.db, "rbac.manage"); err != nil {
		return c.Status(http.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
	}
	uid, err := c.ParamsInt("id")
	if err != nil || uid <= 0 {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user id"})
	}
	matrix, err := h.svc.GetUserRolePermissionMatrix(c.UserContext(), uid)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(matrix)
}

func (h *RBACHandler) GetMatrix(c *fiber.Ctx) error {
	userID, ok := utils.GetUserIDInt(c)
	if !ok || userID <= 0 {
		return client_error.ResponseError(c, fiber.StatusUnauthorized, nil, "Unauthorized")
	}
	matrix, err := h.svc.GetMatrix(c.UserContext())
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, err.Error())
	}
	return c.JSON(matrix)
}

// ---- utils
func parsePaging(c *fiber.Ctx, defLimit, defOffset int) (int, int) {
	limit := utils.GetQueryAsInt(c, "limit", defLimit)
	offset := utils.GetQueryAsInt(c, "offset", defOffset)
	if limit <= 0 {
		limit = defLimit
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}
