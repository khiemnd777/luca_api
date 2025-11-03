package main

import (
	"log"
	"os"

	entsql "entgo.io/ent/dialect/sql"
	gateway "github.com/khiemnd777/andy_api/gateway/runtime"
	"github.com/khiemnd777/andy_api/shared/app/fiber_app"
	"github.com/khiemnd777/andy_api/shared/cache"
	"github.com/khiemnd777/andy_api/shared/circuitbreaker"
	"github.com/khiemnd777/andy_api/shared/config"
	"github.com/khiemnd777/andy_api/shared/cron"
	"github.com/khiemnd777/andy_api/shared/db"
	"github.com/khiemnd777/andy_api/shared/db/ent"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/gen"
	"github.com/khiemnd777/andy_api/shared/logger"
	"github.com/khiemnd777/andy_api/shared/redis"
	"github.com/khiemnd777/andy_api/shared/worker"
	_ "github.com/khiemnd777/andy_api/starter"
)

func main() {
	logger.Init()

	log.Println("🔧 Loading config file...")
	if err := config.Init("config.yaml"); err != nil {
		log.Println("❌ Load config failed:", err)
		os.Exit(1)
	}

	log.Println("✅ Config file loaded!")

	log.Println("🚀 Starting Project...",
		"project:", config.Get().Project.Name,
		"api_prefix:",
		config.Get().Project.BaseAPIPrefix,
		config.Get().Project.Version,
	)

	if err := gen.GenerateFlywayConfig(); err != nil {
		os.Exit(1)
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

	cache.InitTTLConstants()

	// Initialize Redis
	redis.Init()

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

	if err := gen.RunFlyway("migrate"); err != nil {
		os.Exit(1)
	}

	circuitbreaker.Init()

	defer worker.StopAllWorkers()

	cron.StartAllCrons()

	_, fApp := fiber_app.Init()
	if err := gateway.Start(fApp); err != nil {
		log.Fatalf("Gateway error: %v", err)
	}
	select {}
}
