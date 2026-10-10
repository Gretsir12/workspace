package database

import (
	"context"
	"fmt"
	"workspaces/internal/platform/config"

	"github.com/go-pg/pg/v10"
)

func GetConnection(cfg *config.Config) *pg.DB {

	addr := fmt.Sprintf("%s:%d", cfg.DBHost, cfg.DBPort)
	db := pg.Connect(&pg.Options{
		Addr:     addr,
		User:     cfg.DBUser,
		Password: cfg.DBPassword,
		Database: cfg.DBName,
	})
	testConnection(db)

	return db
}

func testConnection(db *pg.DB) {
	ctx := context.Background()

	if err := db.Ping(ctx); err != nil {
		panic(err)
	}
}
