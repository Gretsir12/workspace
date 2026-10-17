package repository

import (
	"context"
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

func (r *WorkspaceRepository) CreateWorkspace(ctx context.Context, workspace *models.Workspace) error {
	_, err := r.db.ModelContext(ctx, workspace).Insert() // ID заполнится через RETURNING
	return err
}

func (r *WorkspaceRepository) FindWorkspaceById(ctx context.Context, id uint32) (*models.Workspace, error) {
	workspace := &models.Workspace{ID: id}
	err := r.db.ModelContext(ctx, workspace).
		Relation("Image").
		WherePK().
		Select()
	if err != nil {
		return nil, err
	}
	return workspace, nil
}

func (r *WorkspaceRepository) UpdateStatus(ctx context.Context, id uint32, status string) error {
	_, err := r.db.ModelContext(ctx, (*models.Workspace)(nil)).
		Set("status = ?", status).
		Where("id = ?", id).
		Update()
	return err
}

func (r *WorkspaceRepository) SetContainer(ctx context.Context, id uint32, containerID, status string) error {
	_, err := r.db.ModelContext(ctx, (*models.Workspace)(nil)).
		Set("container_id = ?, status = ?", containerID, status).
		Where("id = ?", id).
		Update()
	return err
}

func (r *WorkspaceRepository) StopAllActive(ctx context.Context) (int, error) {
	res, err := r.db.ModelContext(ctx, (*models.Workspace)(nil)).
		Set("status = ?", models.WorkspaceStopped).
		Where("status IN (?)", pg.In([]string{models.WorkspacePending, models.WorkspaceRunning})).
		Update()
	if err != nil {
		return 0, err
	}
	return res.RowsAffected(), nil
}
