package repository

import (
	"context"
	"workspaces/internal/platform/config"
	"workspaces/projects/internal/models"

	"github.com/go-pg/pg/v10"
)

type ProjectRepository struct {
	cfg *config.Config
	db  *pg.DB
}

func NewProjectRepository(config *config.Config, db *pg.DB) *ProjectRepository {
	return &ProjectRepository{
		cfg: config,
		db:  db,
	}
}

func (r *ProjectRepository) CreateProject(ctx context.Context, project *models.Project) error {
	_, err := r.db.ModelContext(ctx, project).Insert()
	return err
}

func (r *ProjectRepository) UpdateProject(ctx context.Context, project *models.Project) error {
	_, err := r.db.ModelContext(ctx, project).WherePK().Update()
	return err
}

func (r *ProjectRepository) GetProjectById(ctx context.Context, id uint32) (models.Project, error) {
	project := models.Project{ID: id}
	err := r.db.ModelContext(ctx, &project).WherePK().Select()
	if err != nil {
		return models.Project{}, err
	}
	return project, nil
}
