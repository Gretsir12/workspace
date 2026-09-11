package service

import (
	"errors"
	"fmt"
	"os"
	"workspaces/projects/internal/models"
	"workspaces/projects/internal/repository"
	"workspaces/projects/internal/util"
)

var (
	ErrProjectNotFound      = errors.New("project not found")
	ErrProjectAlreadyExists = errors.New("project already exists")
	ErrInvalidProject       = errors.New("invalid project")
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

func (s *ProjectService) CreateProject(body *models.CreateProjectRequest) error {
	if body == nil || body.Title == nil || *body.Title == "" {
		return fmt.Errorf("create project: %w", ErrInvalidProject)
	}

	if _, err := os.Stat(s.cfg.PROJECTS_PATH + *body.Title); err == nil {
		return fmt.Errorf(
			"create project %q: %w",
			*body.Title,
			ErrProjectAlreadyExists,
		)
	}

	fmt.Println(s.cfg.PROJECTS_PATH + *body.Title)
	if err := os.MkdirAll(s.cfg.PROJECTS_PATH+*body.Title, 0755); err != nil {
		return fmt.Errorf("create project directory: %w", err)
	}

	return nil
}

func (s *ProjectService) buildContainer() {

}
