package repository

import (
	"workspaces/internal/platform/config"
	"workspaces/projects/internal/models"

	"github.com/go-pg/pg/v10"
)

type WorkspaceRepository struct {
	cfg *config.Config
	db  *pg.DB
}

func NewWorkspaceRepository(config *config.Config, db *pg.DB) *WorkspaceRepository {
	return &WorkspaceRepository{
		cfg: config,
		db:  db,
	}
}

func (r *WorkspaceRepository) CreateWorkspace(workspace *models.Workspace) error {
	_, err := r.db.Model(workspace).Insert()
	return err
}
