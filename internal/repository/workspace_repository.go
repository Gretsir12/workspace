package repository

import (
	"workspaces/projects/internal/util"

	"github.com/go-pg/pg/v10"
)

type WorkspaceRepository struct {
	cfg *util.Config
	db  *pg.DB
}

func NewWorkspaceRepository(config *util.Config, db *pg.DB) *WorkspaceRepository {
	return &WorkspaceRepository{
		cfg: config,
		db:  db,
	}
}

func (r *WorkspaceRepository) CreateWorkspace() error {
	return nil
}
