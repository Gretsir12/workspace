package repository

import (
	"workspaces/projects/internal/models"
	"workspaces/projects/internal/util"

	"github.com/go-pg/pg/v10"
)

type ProjectRepository struct {
	cfg *util.Config
	db  *pg.DB
}

func NewProjectRepository(config *util.Config, db *pg.DB) *ProjectRepository {
	return &ProjectRepository{
		cfg: config,
		db:  db,
	}
}

func (r *ProjectRepository) CreateProject(project *models.Project) error {
	_, err := r.db.Model(project).Insert()
	return err
}

func (r *ProjectRepository) GetProjectById(id uint32) (models.Project, error) {
	project := models.Project{ID: id}
	err := r.db.Model(project).WherePK().Select()
	if err != nil {
		return models.Project{}, err
	}
	return project, nil
}
