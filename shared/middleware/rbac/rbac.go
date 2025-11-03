package rbac

import (
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/khiemnd777/andy_api/shared/cache"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/user"
	"github.com/khiemnd777/andy_api/shared/logger"
	"github.com/khiemnd777/andy_api/shared/utils"
)

// ===================================
// Cache versions & key helpers
// ===================================

const (
	roleSetCacheVersion = "v1"
	permSetCacheVersion = "v1"
)

func userRoleSetKey(userID int) string {
	return fmt.Sprintf("user:%d:roles:set:%s", userID, roleSetCacheVersion)
}
func userPermSetKey(userID int) string {
	return fmt.Sprintf("user:%d:permissions:set:%s", userID, permSetCacheVersion)
}

// Invalidate helpers (call these after mutating roles/permissions)
func InvalidateUserRoleSet(userID int)       { cache.InvalidateKeys(userRoleSetKey(userID)) }
func InvalidateUserPermissionSet(userID int) { cache.InvalidateKeys(userPermSetKey(userID)) }

// ==========================
// Role guards (ANY / ALL)
// ==========================

type requireMode int

const (
	anyMode requireMode = iota
	allMode
)

// Backward-compatible: single role -> ANY mode.
func GuardRole(c *fiber.Ctx, roleName string, dbEnt *generated.Client) error {
	return GuardAnyRole(c, dbEnt, roleName)
}

// Require user has ANY of the provided role names (case-insensitive).
func GuardAnyRole(c *fiber.Ctx, dbEnt *generated.Client, roleNames ...string) error {
	return guardRoles(c, dbEnt, anyMode, roleNames...)
}

// Require user has ALL of the provided role names (case-insensitive).
func GuardAllRoles(c *fiber.Ctx, dbEnt *generated.Client, roleNames ...string) error {
	return guardRoles(c, dbEnt, allMode, roleNames...)
}

func guardRoles(c *fiber.Ctx, dbEnt *generated.Client, mode requireMode, roleNames ...string) error {
	uid, ok := utils.GetUserIDInt(c)
	if !ok || uid <= 0 {
		logger.Debug(fmt.Sprintf("GuardRoles: missing/invalid userID; mode=%v roles=%v", mode, roleNames))
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	start := time.Now()
	defer func() {
		logger.Debug(fmt.Sprintf(
			"GuardRoles performance: userID=%v mode=%v roles=%v took=%v",
			uid, mode, roleNames, time.Since(start),
		))
	}()

	ctx := c.UserContext()
	req := normalizeStrings(roleNames)
	if len(req) == 0 {
		logger.Warn("GuardRoles: empty roleNames")
		return c.Status(http.StatusForbidden).JSON(fiber.Map{"error": "Forbidden: no role specified"})
	}

	roleSetPtr, err := cache.Get(userRoleSetKey(uid), cache.TTLLong, func() (*map[string]struct{}, error) {
		// User -> Roles (implicit M2M). Field name is RoleName.
		roles, dbErr := dbEnt.User.
			Query().
			Where(user.IDEQ(uid)).
			QueryRoles().
			All(ctx)
		if dbErr != nil {
			return nil, dbErr
		}
		set := make(map[string]struct{}, len(roles))
		for _, r := range roles {
			if r == nil {
				continue
			}
			name := strings.ToLower(strings.TrimSpace(r.RoleName))
			if name != "" {
				set[name] = struct{}{}
			}
		}
		logger.Debug(fmt.Sprintf("GuardRoles: DB roles userID=%d roles=%v", uid, mapKeys(set)))
		return &set, nil
	})
	if err != nil {
		logger.Error(fmt.Sprintf("GuardRoles: cache/DB error userID=%d err=%v", uid, err))
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "DB error"})
	}
	roleSet := *roleSetPtr

	allowed := false
	missing := make([]string, 0, len(req))
	switch mode {
	case anyMode:
		for _, want := range req {
			if _, ok := roleSet[want]; ok {
				allowed = true
				break
			}
		}
		if !allowed {
			missing = append(missing, req...)
		}
	case allMode:
		allowed = true
		for _, want := range req {
			if _, ok := roleSet[want]; !ok {
				missing = append(missing, want)
				allowed = false
			}
		}
	}

	if !allowed {
		logger.Debug(fmt.Sprintf("GuardRoles forbidden: userID=%d mode=%v have=%v need=%v missing=%v",
			uid, mode, mapKeys(roleSet), req, missing))
		return c.Status(http.StatusForbidden).JSON(fiber.Map{
			"error":   "Forbidden: missing role",
			"details": fiber.Map{"required": req, "missing": missing},
		})
	}
	return nil
}

// ==========================
// Permission guards (ANY/ALL)
// ==========================

// Require user has ANY of the provided permission values (e.g. "product_read").
func GuardAnyPermission(c *fiber.Ctx, dbEnt *generated.Client, permValues ...string) error {
	return guardPermissions(c, dbEnt, anyMode, permValues...)
}

// Require user has ALL of the provided permission values.
func GuardAllPermissions(c *fiber.Ctx, dbEnt *generated.Client, permValues ...string) error {
	return guardPermissions(c, dbEnt, allMode, permValues...)
}

func guardPermissions(c *fiber.Ctx, dbEnt *generated.Client, mode requireMode, permValues ...string) error {
	uid, ok := utils.GetUserIDInt(c)
	logger.Debug(fmt.Sprintf("GuardPermissions: userID=%v mode=%v perms=%v", uid, mode, permValues))
	if !ok || uid <= 0 {
		logger.Debug(fmt.Sprintf("GuardPermissions: missing/invalid userID; mode=%v perms=%v", mode, permValues))
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	start := time.Now()
	defer func() {
		logger.Debug(fmt.Sprintf(
			"GuardPermissions performance: userID=%v mode=%v perms=%v took=%v",
			uid, mode, permValues, time.Since(start),
		))
	}()

	ctx := c.UserContext()
	req := normalizeStrings(permValues)
	if len(req) == 0 {
		logger.Warn("GuardPermissions: empty permValues")
		return c.Status(http.StatusForbidden).JSON(fiber.Map{"error": "Forbidden: no permission specified"})
	}

	permSetPtr, err := cache.Get(userPermSetKey(uid), cache.TTLLong, func() (*map[string]struct{}, error) {
		// User -> Roles -> Permissions (implicit M2M both sides)
		perms, dbErr := dbEnt.User.
			Query().
			Where(user.IDEQ(uid)).
			QueryRoles().
			QueryPermissions().
			All(ctx)
		if dbErr != nil {
			return nil, dbErr
		}
		set := make(map[string]struct{}, len(perms))
		for _, p := range perms {
			if p == nil {
				continue
			}
			val := strings.ToLower(strings.TrimSpace(p.PermissionValue))
			if val != "" {
				set[val] = struct{}{}
			}
		}
		logger.Debug(fmt.Sprintf("GuardPermissions: DB perms userID=%d perms=%v", uid, mapKeys(set)))
		return &set, nil
	})
	if err != nil {
		logger.Error(fmt.Sprintf("GuardPermissions: cache/DB error userID=%d err=%v", uid, err))
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "DB error"})
	}
	permSet := *permSetPtr

	allowed := false
	missing := make([]string, 0, len(req))
	switch mode {
	case anyMode:
		for _, want := range req {
			if _, ok := permSet[want]; ok {
				allowed = true
				break
			}
		}
		if !allowed {
			missing = append(missing, req...)
		}
	case allMode:
		allowed = true
		for _, want := range req {
			if _, ok := permSet[want]; !ok {
				missing = append(missing, want)
				allowed = false
			}
		}
	}

	if !allowed {
		logger.Debug(fmt.Sprintf("GuardPermissions forbidden: userID=%d mode=%v have=%v need=%v missing=%v",
			uid, mode, mapKeys(permSet), req, missing))
		return c.Status(http.StatusForbidden).JSON(fiber.Map{
			"error":   "Forbidden: missing permission",
			"details": fiber.Map{"required": req, "missing": missing},
		})
	}
	return nil
}

// ==========================
// Helpers
// ==========================

func normalizeStrings(in []string) []string {
	out := make([]string, 0, len(in))
	seen := make(map[string]struct{}, len(in))
	for _, s := range in {
		n := strings.ToLower(strings.TrimSpace(s))
		if n == "" {
			continue
		}
		if _, ok := seen[n]; ok {
			continue
		}
		seen[n] = struct{}{}
		out = append(out, n)
	}
	slices.Sort(out)
	return out
}

func mapKeys(m map[string]struct{}) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	slices.Sort(out)
	return out
}
