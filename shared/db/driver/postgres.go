package driver

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/khiemnd777/andy_api/shared/config"
	dbinterface "github.com/khiemnd777/andy_api/shared/db/interface"
	_ "github.com/lib/pq"
)

type PostgresClient struct {
	DB     *sql.DB
	Config config.PostgresConfig
}

func NewPostgresClient(cfg config.PostgresConfig) dbinterface.DatabaseClient {
	return &PostgresClient{Config: cfg}
}

func (p *PostgresClient) Connect() error {
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		p.Config.Host, p.Config.Port, p.Config.User, p.Config.Password, p.Config.Name, p.Config.SSLMode,
	)
	db, err := sql.Open("postgres", connStr)
	if err == nil && db.Ping() == nil {
		p.DB = db
		return nil
	}

	// Check if error is "database does not exist"
	if err != nil || db.Ping() != nil {
		// Connect to default "postgres" database
		defaultConnStr := fmt.Sprintf(
			"host=%s port=%d user=%s password=%s dbname=postgres sslmode=%s",
			p.Config.Host, p.Config.Port, p.Config.User, p.Config.Password, p.Config.SSLMode,
		)
		defaultDB, err := sql.Open("postgres", defaultConnStr)
		if err != nil {
			return fmt.Errorf("failed to connect to default postgres DB: %w", err)
		}
		defer defaultDB.Close()

		// Create target database if not exists
		createSQL := fmt.Sprintf(`CREATE DATABASE "%s"`, p.Config.Name)
		if _, err := defaultDB.Exec(createSQL); err != nil {
			// If already exists, ignore
			if !strings.Contains(err.Error(), "already exists") {
				return fmt.Errorf("failed to create database: %w", err)
			}
		}
	}

	// Try reconnecting to the target DB again
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		return err
	}
	p.DB = db
	return db.Ping()
}

func (p *PostgresClient) Close() error {
	if p.DB != nil {
		return p.DB.Close()
	}
	return nil
}

func (p *PostgresClient) GetSQL() *sql.DB {
	return p.DB
}
