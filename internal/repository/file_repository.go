package repository

import (
	"workspaces/projects/internal/util"

	"github.com/go-pg/pg/v10"
)

type FileRepository struct {
	cfg *util.Config
	db  *pg.DB
}

func NewFileRepository(config *util.Config, db *pg.DB) *FileRepository {
	return &FileRepository{
		cfg: config,
		db:  db,
	}
}
