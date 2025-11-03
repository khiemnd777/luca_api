package ent

import (
	"database/sql"
	"fmt"

	entsql "entgo.io/ent/dialect/sql"
)

type EntClientBuilder func(drv *entsql.Driver) any

var entHandlers = map[string]func(db *sql.DB, builder EntClientBuilder) (any, error){
	"postgres": func(db *sql.DB, builder EntClientBuilder) (any, error) {
		drv := entsql.OpenDB("postgres", db)
		return builder(drv), nil
	},
	// "mongodb", "sqllite", v.v...
}

func InitEntClient(provider string, db *sql.DB, builder EntClientBuilder) (any, error) {
	handler, ok := entHandlers[provider]
	if !ok {
		return nil, fmt.Errorf("no Ent client initializer for provider: %s", provider)
	}
	return handler(db, builder)
}
