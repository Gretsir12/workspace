package repository

import (
	"workspaces/projects/internal/models"
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

func (r *WorkspaceRepository) CreateWorkspace(workspace *models.Workspace) error {
	_, err := r.db.Model(workspace).Insert()
	return err
}
