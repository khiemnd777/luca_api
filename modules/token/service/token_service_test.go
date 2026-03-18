package service

import "testing"

func TestUserPermissionCacheKeyIncludesUserID(t *testing.T) {
	if got := userPermissionCacheKey(1); got != "user:1:perms" {
		t.Fatalf("unexpected permission cache key: %s", got)
	}
	if userPermissionCacheKey(1) == userPermissionCacheKey(2) {
		t.Fatal("permission cache keys should differ per user")
	}
}

func TestUserDepartmentCacheKeyIncludesUserID(t *testing.T) {
	if got := userDepartmentCacheKey(1); got != "user:1:dept" {
		t.Fatalf("unexpected department cache key: %s", got)
	}
	if userDepartmentCacheKey(1) == userDepartmentCacheKey(2) {
		t.Fatal("department cache keys should differ per user")
	}
}
