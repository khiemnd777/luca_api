package driver

import (
	"context"
	"database/sql"
	"time"

	"github.com/khiemnd777/andy_api/shared/config"
	dbinterface "github.com/khiemnd777/andy_api/shared/db/interface"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoClient struct {
	Client   *mongo.Client
	Database *mongo.Database
	Config   config.MongoConfig
}

func NewMongoClient(cfg config.MongoConfig) dbinterface.DatabaseClient {
	return &MongoClient{Config: cfg}
}

func (m *MongoClient) Connect() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(m.Config.URI))
	if err != nil {
		return err
	}
	m.Client = client
	m.Database = client.Database(m.Config.Database)
	return nil
}

func (m *MongoClient) Close() error {
	if m.Client != nil {
		return m.Client.Disconnect(context.Background())
	}
	return nil
}

func (m *MongoClient) GetSQL() *sql.DB {
	return nil // Mongo không có SQL
}
