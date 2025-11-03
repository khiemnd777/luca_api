package service

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/khiemnd777/andy_api/modules/rbac/repository"
	"github.com/khiemnd777/andy_api/shared/cache"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/logger"
	"github.com/khiemnd777/andy_api/shared/middleware/rbac"
)

type ListResult[T any] struct {
	Items []T `json:"items"`
	Total int `json:"total"`
}

type RBACService interface {
	// Role
	CreateRole(ctx context.Context, name, displayName, brief string) (int, error)
	RenameRole(ctx context.Context, roleID int, newName string) error
	UpdateRole(ctx context.Context, roleID int, newName, newDisplayName, newBrief string) error
	DeleteRole(ctx context.Context, roleID int) error
	ListRoles(ctx context.Context, limit, offset int) (*ListResult[*generated.Role], error)
	ListUserRoles(ctx context.Context, userID, limit, offset int) ([]*generated.Role, error)

	// Permission
	CreatePermission(ctx context.Context, name, value string) (int, error)
	DeletePermission(ctx context.Context, id int) error
	ListPermissions(ctx context.Context, limit, offset int) (*ListResult[*generated.Permission], error)

	// Matrix (role-permission)
	ReplaceRolePermissions(ctx context.Context, roleID int, permIDs []int) error
	AddRolePermissions(ctx context.Context, roleID int, permIDs []int) error
	RemoveRolePermissions(ctx context.Context, roleID int, permIDs []int) error
	GetRolePermissionIDs(ctx context.Context, roleID int) ([]int, error)
	GetRolePermissionMatrix(ctx context.Context, roleIDs []int) (map[int][]int, error)
	GetUserPermissionIDs(ctx context.Context, userID int) ([]int, error)
	GetUserRolePermissionMatrix(ctx context.Context, userID int) (*RolePermissionMatrix, error)
}

// -------- Cache
const (
	ttlShort  = 30 * time.Second
	ttlMedium = 5 * time.Minute
	ttlLong   = 30 * time.Minute
)
const (
	kRoleListAll = "rbac:roles:list:*"
	kPermListAll = "rbac:perms:list:*"
)

func kRoleList(limit, offset int) string { return fmt.Sprintf("rbac:roles:list:%d:%d", limit, offset) }
func kPermList(limit, offset int) string { return fmt.Sprintf("rbac:perms:list:%d:%d", limit, offset) }
func kUserRoles(userID, limit, offset int) string {
	return fmt.Sprintf("user:%d:rbac:roles:%d:%d", userID, limit, offset)
}
func kUserMatrix(userID int) string { return fmt.Sprintf("user:%d:rbac:matrix", userID) }

type rbacService struct {
	roles RoleRepo
	perms PermRepo
}

type RoleRepo = repository.RoleRepository
type PermRepo = repository.PermissionRepository

func NewRBACService(roles RoleRepo, perms PermRepo) RBACService {
	return &rbacService{roles: roles, perms: perms}
}

// -------- Role
func (s *rbacService) CreateRole(ctx context.Context, name, displayName, brief string) (int, error) {
	name = strings.TrimSpace(name)
	r, err := s.roles.Create(ctx, name, displayName, brief)
	if err != nil {
		return 0, err
	}
	cache.InvalidateKeys(kRoleListAll)
	return r.ID, nil
}
func (s *rbacService) RenameRole(ctx context.Context, roleID int, newName string) error {
	_, err := s.roles.UpdateName(ctx, roleID, strings.TrimSpace(newName))
	if err == nil {
		cache.InvalidateKeys(kRoleListAll)
	}
	return err
}
func (s *rbacService) UpdateRole(ctx context.Context, roleID int, newName, newDisplayName, newBrief string) error {
	_, err := s.roles.Update(ctx, roleID, strings.TrimSpace(newName), strings.TrimSpace(newDisplayName), strings.TrimSpace(newBrief))
	if err == nil {
		cache.InvalidateKeys(kRoleListAll)
	}
	return err
}
func (s *rbacService) DeleteRole(ctx context.Context, roleID int) error {
	// Invalidate affected users
	userIDs, _ := s.roles.UserIDsOfRole(ctx, roleID)
	if err := s.roles.Delete(ctx, roleID); err != nil {
		return err
	}
	for _, uid := range userIDs {
		rbac.InvalidateUserRoleSet(uid)
		rbac.InvalidateUserPermissionSet(uid)
		cache.InvalidateKeys(kUserMatrix(uid))
	}
	cache.InvalidateKeys(kRoleListAll)
	return nil
}

func (s *rbacService) ListRoles(ctx context.Context, limit, offset int) (*ListResult[*generated.Role], error) {
	key := kRoleList(limit, offset)
	res, err := cache.Get(key, ttlMedium, func() (*ListResult[*generated.Role], error) {
		items, total, err := s.roles.List(ctx, limit, offset)
		if err != nil {
			return nil, err
		}
		return &ListResult[*generated.Role]{Items: items, Total: total}, nil
	})
	if err != nil {
		return nil, err
	}
	return res, nil
}
func (s *rbacService) ListUserRoles(ctx context.Context, userID, limit, offset int) ([]*generated.Role, error) {
	key := kUserRoles(userID, limit, offset)
	return cache.GetList(key, ttlMedium, func() ([]*generated.Role, error) {
		items, _, err := s.roles.ListByUser(ctx, userID, limit, offset)
		return items, err
	})
}

// -------- Permission
func (s *rbacService) CreatePermission(ctx context.Context, name, value string) (int, error) {
	name = strings.TrimSpace(name)
	value = strings.ToLower(strings.TrimSpace(value))
	p, err := s.perms.Create(ctx, name, value)
	if err != nil {
		return 0, err
	}
	cache.InvalidateKeys(kPermListAll)
	return p.ID, nil
}
func (s *rbacService) DeletePermission(ctx context.Context, id int) error {
	// Find roles affected to invalidate users
	// (optional) we could query backward, but simple path: delete -> invalidate all users of roles that contained it.
	// For scale, prefetch roles->users before delete if needed.
	err := s.perms.Delete(ctx, id)
	if err == nil {
		cache.InvalidateKeys(kPermListAll)
	}
	return err
}
func (s *rbacService) ListPermissions(ctx context.Context, limit, offset int) (*ListResult[*generated.Permission], error) {
	key := kPermList(limit, offset)
	res, err := cache.Get(key, ttlLong, func() (*ListResult[*generated.Permission], error) {
		items, total, err := s.perms.List(ctx, limit, offset)
		if err != nil {
			return nil, err
		}
		return &ListResult[*generated.Permission]{Items: items, Total: total}, nil
	})
	if err != nil {
		return nil, err
	}
	return res, nil
}

// -------- Matrix
type PermissionMeta struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Value string `json:"value"`
}

type MatrixRow struct {
	RoleID   int    `json:"role_id"`
	RoleName string `json:"role_name"`
	Flags    []bool `json:"flags"` // theo thứ tự PermissionIDs
}

type RolePermissionMatrix struct {
	Permissions []PermissionMeta `json:"permissions"`
	Roles       []MatrixRow      `json:"roles"`
}

func (s *rbacService) ReplaceRolePermissions(ctx context.Context, roleID int, permIDs []int) error {
	if err := s.roles.ReplacePermissions(ctx, roleID, permIDs); err != nil {
		return err
	}
	s.invalidateUsersOfRole(ctx, roleID)
	return nil
}
func (s *rbacService) AddRolePermissions(ctx context.Context, roleID int, permIDs []int) error {
	if err := s.roles.AddPermissions(ctx, roleID, permIDs); err != nil {
		return err
	}
	s.invalidateUsersOfRole(ctx, roleID)
	return nil
}
func (s *rbacService) RemoveRolePermissions(ctx context.Context, roleID int, permIDs []int) error {
	if err := s.roles.RemovePermissions(ctx, roleID, permIDs); err != nil {
		return err
	}
	s.invalidateUsersOfRole(ctx, roleID)
	return nil
}

func (s *rbacService) invalidateUsersOfRole(ctx context.Context, roleID int) {
	uids, err := s.roles.UserIDsOfRole(ctx, roleID)
	if err != nil {
		logger.Warn("RBAC invalidateUsersOfRole fetch failed: " + err.Error())
		return
	}
	for _, uid := range uids {
		rbac.InvalidateUserPermissionSet(uid)
		cache.InvalidateKeys(kUserMatrix(uid))
	}
}

func (s *rbacService) GetRolePermissionIDs(ctx context.Context, roleID int) ([]int, error) {
	ids, err := s.roles.PermissionIDsOfRole(ctx, roleID)
	return ids, err
}

func (s *rbacService) GetRolePermissionMatrix(ctx context.Context, roleIDs []int) (map[int][]int, error) {
	out := make(map[int][]int, len(roleIDs))
	for _, rid := range roleIDs {
		ids, err := s.roles.PermissionIDsOfRole(ctx, rid)
		if err != nil {
			return nil, err
		}
		out[rid] = ids
	}
	return out, nil
}
func (s *rbacService) GetUserPermissionIDs(ctx context.Context, userID int) ([]int, error) {
	// Lấy tất cả role của user
	roles, _, err := s.roles.ListByUser(ctx, userID, 1000, 0)
	if err != nil {
		return nil, err
	}
	// Hợp nhất tất cả permission IDs từ các role (de-dup + sort)
	set := make(map[int]struct{}, 64)
	for _, r := range roles {
		pids, err := s.roles.PermissionIDsOfRole(ctx, r.ID)
		if err != nil {
			return nil, err
		}
		for _, id := range pids {
			set[id] = struct{}{}
		}
	}
	out := make([]int, 0, len(set))
	for id := range set {
		out = append(out, id)
	}
	sort.Ints(out)
	return out, nil
}
func (s *rbacService) GetUserRolePermissionMatrix(ctx context.Context, userID int) (*RolePermissionMatrix, error) {
	key := kUserMatrix(userID)
	return cache.Get(key, ttlMedium, func() (*RolePermissionMatrix, error) {
		// 1️⃣ Lấy toàn bộ permissions (để có cột ổn định)
		perms, _, err := s.perms.List(ctx, 10000, 0)
		if err != nil {
			return nil, err
		}

		permissionMetas := make([]PermissionMeta, len(perms))
		indexByPermID := make(map[int]int, len(perms))
		for i, p := range perms {
			permissionMetas[i] = PermissionMeta{
				ID:    p.ID,
				Name:  p.PermissionName,
				Value: p.PermissionValue,
			}
			indexByPermID[p.ID] = i
		}

		// 2️⃣ Lấy danh sách role của user
		roles, _, err := s.roles.ListByUser(ctx, userID, 10000, 0)
		if err != nil {
			return nil, err
		}

		// 3️⃣ Tạo từng dòng (Role × Permission)
		rows := make([]MatrixRow, 0, len(roles))
		for _, r := range roles {
			flags := make([]bool, len(permissionMetas))
			pids, err := s.roles.PermissionIDsOfRole(ctx, r.ID)
			if err != nil {
				return nil, err
			}
			for _, pid := range pids {
				if col, ok := indexByPermID[pid]; ok {
					flags[col] = true
				}
			}
			rows = append(rows, MatrixRow{
				RoleID:   r.ID,
				RoleName: r.RoleName,
				Flags:    flags,
			})
		}

		return &RolePermissionMatrix{
			Permissions: permissionMetas,
			Roles:       rows,
		}, nil
	})
}
