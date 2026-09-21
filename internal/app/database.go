package app

import (
	"context"
	"strconv"
	"workspaces/projects/internal/models"
	"workspaces/projects/internal/util"

	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
)

func GetConnection(cfg *util.Config) *pg.DB {

	// fmt.Println(strconv.FormatUint(uint64(cfg.DATABASE_PORT), 10))
	// fmt.Println(cfg.DATABASE_PORT)
	db := pg.Connect(&pg.Options{
		Addr:     ":" + strconv.FormatUint(uint64(cfg.DATABASE_PORT), 10),
		User:     cfg.DATABASE_USER,
		Password: cfg.DATABASE_PASSWORD,
		Database: cfg.DATABASE_DB,
	})
	testConnection(db)
	createSchema(db)

	return db
}

func testConnection(db *pg.DB) {
	ctx := context.Background()

	if err := db.Ping(ctx); err != nil {
		panic(err)
	}
}

func createSchema(db *pg.DB) error {
	models := []interface{}{
		(*models.Project)(nil),
		(*models.File)(nil),
	}

	for _, model := range models {
		err := db.Model(model).CreateTable(&orm.CreateTableOptions{
			Temp: false,
		})
		if err != nil {
			return err
		}
	}
	return nil
}
