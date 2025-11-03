package main

import (
	"fmt"
	"log"
	"os"

	entsql "entgo.io/ent/dialect/sql"

	"github.com/khiemnd777/andy_api/shared/config"
	"github.com/khiemnd777/andy_api/shared/db"
	"github.com/khiemnd777/andy_api/shared/db/ent"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/gen"
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
	dbClient, err := db.NewDatabaseClient(dbCfg)
	if err != nil {
		log.Fatalf("Cannot initialize DB: %v", err)
	}
	defer dbClient.Close()

	if err := dbClient.Connect(); err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}
	log.Println("Connected to DB successfully!")

	if err := gen.GenerateEntClient(); err != nil {
		os.Exit(1)
	}

	sqlDB := dbClient.GetSQL() // Returns *sql.DB if Postgres, but nil Mongo

	_, entErr := ent.EntBootstrap(dbCfg.Provider, sqlDB, func(drv *entsql.Driver) any {
		return generated.NewClient(generated.Driver(drv))
	}, dbCfg.AutoMigrate)
	if entErr != nil {
		log.Fatalf("❌ Failed to init Ent client: %v", entErr)
		os.Exit(1)
	}
}
