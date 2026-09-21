package service

import (
	"fmt"
	"os"
	"strconv"
	"workspaces/projects/internal/models"
	"workspaces/projects/internal/repository"
	"workspaces/projects/internal/util"
)

type ProjectService struct {
	r   *repository.ProjectRepository
	cfg *util.Config
}

func NewProjectService(config *util.Config, r *repository.ProjectRepository) *ProjectService {
	return &ProjectService{
		r:   r,
		cfg: config,
	}
}

func (s *ProjectService) CreateProject(body *models.CreateProjectRequest) (models.Project, error) {
	if body == nil || body.Title == nil || *body.Title == "" {
		return models.Project{}, fmt.Errorf("create project: %w", util.ErrInvalidProject)
	}

	// creating project object and saving it to database
	project := models.Project{
		User:  12,
		Title: *body.Title,
		Files: []models.File{},
	}
	err := s.r.CreateProject(&project)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(s.cfg.PROJECTS_PATH + strconv.FormatUint(uint64(project.ID), 10))
	if err := os.MkdirAll(s.cfg.PROJECTS_PATH+strconv.FormatUint(uint64(project.ID), 10)+"/"+project.Title, 0755); err != nil {
		return project, fmt.Errorf("create project directory: %w", err)
	}

	return project, nil
}

func (s *ProjectService) buildContainer() {

}
