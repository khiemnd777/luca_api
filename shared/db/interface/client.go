package dbinterface

import "database/sql"

type DatabaseClient interface {
	Connect() error
	Close() error
	GetSQL() *sql.DB // Optional for relational DB
}
