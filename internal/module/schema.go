package module

import (
	"workspaces/projects/internal/models"

	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
)

func createSchema(db *pg.DB) error {
	models := []interface{}{
		(*models.Project)(nil),
		(*models.File)(nil),
		(*models.Workspace)(nil),
	}

	for _, model := range models {
		err := db.Model(model).CreateTable(&orm.CreateTableOptions{
			Temp:          false,
			FKConstraints: true,
			IfNotExists:   true,
		})
		if err != nil {
			return err
		}
	}
	return nil
}
