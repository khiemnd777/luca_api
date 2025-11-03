package db

// import (
// 	"database/sql"
// 	"fmt"
// 	"log"

// 	"github.com/khiemnd777/andy_api/shared/config"
// 	_ "github.com/lib/pq"
// )

// func InitDatabase(dbCfg config.DatabaseConfig) (*sql.DB, error) {
// 	// Kết nối tạm đến database "postgres" để kiểm tra database chính có tồn tại chưa
// 	tempDSN := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=postgres sslmode=%s",
// 		dbCfg.Host, dbCfg.Port, dbCfg.User, dbCfg.Password, dbCfg.SSLMode)
// 	tempDB, err := sql.Open("postgres", tempDSN)
// 	if err != nil {
// 		return nil, fmt.Errorf("connect to temp db failed: %w", err)
// 	}
// 	defer tempDB.Close()

// 	// Kiểm tra database đã tồn tại hay chưa
// 	var exists bool
// 	query := fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = '%s')", dbCfg.Name)
// 	if err := tempDB.QueryRow(query).Scan(&exists); err != nil {
// 		return nil, fmt.Errorf("check db existence failed: %w", err)
// 	}

// 	if !exists {
// 		log.Printf("🔧 Creating database '%s'...\n", dbCfg.Name)
// 		if _, err := tempDB.Exec(fmt.Sprintf("CREATE DATABASE %s", dbCfg.Name)); err != nil {
// 			return nil, fmt.Errorf("create database failed: %w", err)
// 		}
// 		log.Printf("✅ Database '%s' created.", dbCfg.Name)
// 	} else {
// 		log.Printf("✅ Database '%s' already exists.", dbCfg.Name)
// 	}

// 	// Kết nối thực sự tới database chính
// 	mainDSN := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
// 		dbCfg.Host, dbCfg.Port, dbCfg.User, dbCfg.Password, dbCfg.Name, dbCfg.SSLMode)
// 	mainDB, err := sql.Open("postgres", mainDSN)
// 	if err != nil {
// 		return nil, fmt.Errorf("connect to main db failed: %w", err)
// 	}

// 	return mainDB, nil
// }
