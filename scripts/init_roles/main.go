package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/khiemnd777/andy_api/shared/config"
	"github.com/khiemnd777/andy_api/shared/utils"
	_ "github.com/lib/pq"
)

const (
	observabilityPermissionName  = "System Log Reader"
	observabilityPermissionValue = "system_log.read"
)

type Role struct {
	RoleName    string
	DisplayName string
	Brief       string
}

func main() {
	cfgerr := config.Init(utils.GetFullPath("config.yaml"))
	if cfgerr != nil {
		panic(fmt.Sprintf("❌ Config not initialized: %v", cfgerr))
	}

	dbCfg := config.Get().Database
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		dbCfg.Postgres.Host, dbCfg.Postgres.Port, dbCfg.Postgres.User, dbCfg.Postgres.Password, dbCfg.Postgres.Name, dbCfg.Postgres.SSLMode,
	)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("❌ Cannot connect DB: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	roles := []Role{
		{RoleName: "user", DisplayName: "User", Brief: "A normal user"},
		{RoleName: "admin", DisplayName: "Administrator", Brief: "Administrator"},
		{RoleName: "guest", DisplayName: "Guest", Brief: "Guest user with limited access"},
	}

	for _, role := range roles {
		roleID, err := ensureRole(ctx, db, role)
		if err != nil {
			log.Fatalf("❌ Failed to ensure role %q: %v", role.RoleName, err)
		}
		fmt.Printf("✅ Ensured role '%s' successfully (id=%d).\n", role.RoleName, roleID)
	}

	permID, err := ensurePermission(ctx, db, observabilityPermissionName, observabilityPermissionValue)
	if err != nil {
		log.Fatalf("❌ Failed to ensure permission %q: %v", observabilityPermissionValue, err)
	}
	adminRoleID, err := getRoleIDByName(ctx, db, "admin")
	if err != nil {
		log.Fatalf("❌ Failed to resolve admin role: %v", err)
	}
	if err := attachPermission(ctx, db, adminRoleID, permID); err != nil {
		log.Fatalf("❌ Failed to assign permission %q to admin role: %v", observabilityPermissionValue, err)
	}
}

func ensureRole(ctx context.Context, db *sql.DB, role Role) (int, error) {
	var id int
	err := db.QueryRowContext(ctx, `
		INSERT INTO roles (role_name, display_name, brief)
		VALUES ($1, NULLIF($2, ''), NULLIF($3, ''))
		ON CONFLICT (role_name) DO UPDATE
		SET display_name = EXCLUDED.display_name,
		    brief = EXCLUDED.brief
		RETURNING id
	`, role.RoleName, role.DisplayName, role.Brief).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func ensurePermission(ctx context.Context, db *sql.DB, name, value string) (int, error) {
	var id int
	err := db.QueryRowContext(ctx, `
		INSERT INTO permissions (permission_name, permission_value)
		VALUES ($1, $2)
		ON CONFLICT (permission_value) DO UPDATE
		SET permission_name = EXCLUDED.permission_name
		RETURNING id
	`, name, value).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func getRoleIDByName(ctx context.Context, db *sql.DB, roleName string) (int, error) {
	var id int
	err := db.QueryRowContext(ctx, `
		SELECT id
		FROM roles
		WHERE role_name = $1
	`, roleName).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func attachPermission(ctx context.Context, db *sql.DB, roleID, permID int) error {
	_, err := db.ExecContext(ctx, `
		INSERT INTO role_permissions (role_id, permission_id)
		VALUES ($1, $2)
		ON CONFLICT (role_id, permission_id) DO NOTHING
	`, roleID, permID)
	return err
}
