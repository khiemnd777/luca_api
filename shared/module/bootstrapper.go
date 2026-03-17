package module

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	appLogger "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/khiemnd777/andy_api/shared/app"
	"github.com/khiemnd777/andy_api/shared/app/fiber_app"
	"github.com/khiemnd777/andy_api/shared/cache"
	"github.com/khiemnd777/andy_api/shared/circuitbreaker"
	"github.com/khiemnd777/andy_api/shared/config"
	"github.com/khiemnd777/andy_api/shared/cron"
	"github.com/khiemnd777/andy_api/shared/db"
	"github.com/khiemnd777/andy_api/shared/logger"
	"github.com/khiemnd777/andy_api/shared/redis"
	"github.com/khiemnd777/andy_api/shared/runtime"
	"github.com/khiemnd777/andy_api/shared/utils"
)

type ModuleDeps[T any] struct {
	Config    *T
	DB        *sql.DB
	Ent       any
	SharedEnt any
	App       *fiber.App
}

type ModuleOptions[T any] struct {
	ConfigPath          string
	ModuleName          string
	InitEntClient       func(provider string, db *sql.DB, cfg *T) (any, error)
	InitSharedEntClient func(provider string, db *sql.DB, cfg *T) (any, error)
	OnRegistry          func(app *fiber.App, deps *ModuleDeps[T])
	OnReady             func(deps *ModuleDeps[T])
}

func StartModule[T any](opts ModuleOptions[T]) {
	logger.Init()
	logger.SetComponent(opts.ModuleName)

	config.Init(utils.GetFullPath("config.yaml"))
	logger.Configure(logger.Options{
		ServiceName:  config.Get().Observability.ServiceName,
		Environment:  config.Get().Observability.Environment,
		Level:        config.Get().Observability.Logs.Level,
		RedactFields: config.Get().Observability.Logs.RedactFields,
		Component:    opts.ModuleName,
	})

	cache.InitTTLConstants()

	redis.Init()

	circuitbreaker.Init()

	logger.Info(fmt.Sprintf("🔧 Starting module: %s", opts.ModuleName))

	// Step 1: Load config
	cfg, err := utils.LoadConfig[T](opts.ConfigPath)
	if err != nil {
		logger.Error(fmt.Sprintf("❌ Failed to load module config: %v", err))
	}

	// Should use `go run scripts/module_runner status` instead.
	// srvCfg := any(cfg).(interface{ GetServer() config.ServerConfig }).GetServer()
	// monitor.InitModuleLifecycle(opts.ModuleName, srvCfg.Port)

	dbCfg := any(cfg).(interface{ GetDatabase() config.DatabaseConfig }).GetDatabase()

	// Step 2: Create DB client
	dbClient, err := db.NewDatabaseClient(dbCfg)
	if err != nil {
		logger.Error(fmt.Sprintf("❌ Cannot create database client: %v", err))
	}
	defer dbClient.Close()

	if err := dbClient.Connect(); err != nil {
		logger.Error(fmt.Sprintf("❌ Failed to connect to database: %v", err))
	}

	// Step 3: Init Ent client
	var entClient any
	sqlDB := dbClient.GetSQL()
	if opts.InitEntClient != nil {
		entClient, err = opts.InitEntClient(any(cfg).(interface{ GetDatabase() config.DatabaseConfig }).GetDatabase().Provider, sqlDB, cfg)
		if err != nil {
			logger.Error(fmt.Sprintf("❌ Failed to init Ent client: %v", err))
		}
	}

	// Step 3.1: Init Shared Ent client if provided
	var sharedEntClient any
	if opts.InitSharedEntClient != nil {
		sharedEntClient, err = (opts.InitSharedEntClient)(any(cfg).(interface{ GetSharedDatabase() config.DatabaseConfig }).GetSharedDatabase().Provider, sqlDB, cfg)
		if err != nil {
			logger.Error(fmt.Sprintf("❌ Failed to init Shared Ent client: %v", err))
		}
	}

	// Step 4: Init Fiber app
	fiberApp := fiber_app.NewFiberApp()
	fiberApp.Use(appLogger.New())

	fiberApp.Get("/health", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	// Step 5: Register routes
	deps := &ModuleDeps[T]{
		Config:    cfg,
		DB:        sqlDB,
		Ent:       entClient,
		SharedEnt: sharedEntClient,
		App:       fiberApp,
	}
	opts.OnRegistry(fiberApp, deps)

	if opts.OnReady != nil {
		opts.OnReady(deps)
	}

	cron.StartAllCrons()

	// Step 6: Start server
	StartFiber(fiberApp, opts.ModuleName)
}

func getDestPort(port int) int {
	mPort := config.Get().Server.Port
	return mPort + port
}

func StartFiber(fiberApp *fiber.App, moduleName string) {
	// 1) Lấy entry của chính module trong tmp/runtime.json
	reg, err := runtime.LoadRegistry()
	if err != nil {
		logger.Error(fmt.Sprintf("cannot load runtime registry: %v", err))
		return
	}
	rm, ok := reg[moduleName]
	if !ok || rm.Host == "" || rm.Port == 0 {
		logger.Error(fmt.Sprintf("runtime entry for [%s] not found or invalid", moduleName))
		return
	}

	host, port := rm.Host, rm.Port
	// destPort := getDestPort(port)
	addr := fmt.Sprintf("%s:%d", host, port)

	// 2) Bind đúng cổng (KHÔNG ListenOnAvailablePort nữa)
	reserved, err := app.ListenOnAvailablePort(host, port)

	if err != nil {
		logger.Error(fmt.Sprintf("❌ Cannot start listener: %v", err))
		return
	}

	// 3) Cập nhật lại runtime (PID + RunAt; port giữ nguyên)
	rm.PID = os.Getpid()
	rm.RunAt = time.Now()
	reg[moduleName] = rm
	_ = runtime.SaveRegistry(reg) // lỗi ghi file không chặn server chạy

	logger.Info(fmt.Sprintf("✅ %s module listening on %s", moduleName, addr))

	// 4) Serve Fiber
	if err := fiberApp.Listener(reserved.Listener); err != nil {
		logger.Error(fmt.Sprintf("❌ Fiber app failed: %v", err))
	}
}
