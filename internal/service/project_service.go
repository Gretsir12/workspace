package service

import (
	"fmt"
	"os"
	"strconv"
	"workspaces/projects/internal/custom_errors"
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
	if err := body.Validate(); err != nil {
		return models.Project{}, fmt.Errorf("create project: %w", custom_errors.ErrInvalidProject)
	}

	// creating project object and saving it to database
	project := models.Project{
		User:         12,
		Title:        *body.Title,
		Files:        []models.File{},
		Dependencies: body.Dependencies,
	}

	if err := s.r.CreateProject(&project); err != nil {
		fmt.Println(err)
	}

	if err := s.createProjectDir(&project); err != nil {
		return project, fmt.Errorf("create project dir: %w", custom_errors.ErrServerError)
	}

	// updating project path

	if err := s.updateProjectPath(&project); err != nil {
		return project, fmt.Errorf("update project dir: %w", custom_errors.ErrServerError)
	}

	return project, nil
}

func (s *ProjectService) createProjectDir(project *models.Project) error {
	fmt.Println(s.cfg.PROJECTS_PATH + strconv.FormatUint(uint64(project.ID), 10) + "/" + project.Title)
	if err := os.MkdirAll(s.cfg.PROJECTS_PATH+strconv.FormatUint(uint64(project.ID), 10)+"/"+project.Title, 0755); err != nil {
		return fmt.Errorf("create project directory: %w", err)
	}
	return nil
}

func (s *ProjectService) updateProjectPath(project *models.Project) error {
	project.Path = s.cfg.PROJECTS_PATH + strconv.FormatUint(uint64(project.ID), 10) + "/" + project.Title

	if err := s.r.UpdateProject(project); err != nil {
		return fmt.Errorf("update path to project directory: %w", custom_errors.ErrProjectNotFound)
	}
	return nil
}

func (s *ProjectService) buildContainer() {

}
