package db

import (
	"fmt"

	"github.com/khiemnd777/andy_api/shared/config"
	"github.com/khiemnd777/andy_api/shared/db/driver"
	dbinterface "github.com/khiemnd777/andy_api/shared/db/interface"
)

func NewDatabaseClient(cfg config.DatabaseConfig) (dbinterface.DatabaseClient, error) {
	switch cfg.Provider {
	case "postgres":
		return driver.NewPostgresClient(cfg.Postgres), nil
	case "mongodb":
		return driver.NewMongoClient(cfg.MongoDB), nil
	default:
		return nil, fmt.Errorf("unsupported database provider: %s", cfg.Provider)
	}
}
