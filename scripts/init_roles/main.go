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

type Role struct {
	ID          int
	Name        string
	Description string
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
		{ID: 1, Name: "user", Description: "A normal user"},
		{ID: 2, Name: "admin", Description: "Administrator"},
		{ID: 3, Name: "guest", Description: "Guest user with limited access"},
	}

	for _, role := range roles {
		var exists bool
		err := db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM roles WHERE id = $1)`, role.ID).Scan(&exists)
		if err != nil {
			log.Fatalf("❌ Failed to check existing role %q: %v", role.Name, err)
		}
		if exists {
			fmt.Printf("✅ Role '%s' already exists. Skipping insert.\n", role.Name)
			continue
		}

		_, err = db.ExecContext(ctx,
			`INSERT INTO roles (id, name, description) VALUES ($1, $2, $3)`,
			role.ID, role.Name, role.Description)
		if err != nil {
			log.Fatalf("❌ Failed to insert role '%s': %v", role.Name, err)
		}
		fmt.Printf("✅ Inserted role '%s' successfully.\n", role.Name)
	}
}
