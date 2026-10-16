package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"workspaces/internal/platform/config"
	"workspaces/projects/internal/custom_errors"
	"workspaces/projects/internal/models"
	"workspaces/projects/internal/repository"

	"github.com/go-pg/pg/v10"
)

type ProjectService struct {
	r   *repository.ProjectRepository
	cfg *config.Config
}

func NewProjectService(config *config.Config, r *repository.ProjectRepository) *ProjectService {
	return &ProjectService{
		r:   r,
		cfg: config,
	}
}

func (s *ProjectService) GetProject(ctx context.Context, projectId uint32) (models.Project, error) {
	project, err := s.r.GetProjectById(ctx, projectId)
	if err != nil {
		if errors.Is(err, pg.ErrNoRows) {
			return models.Project{}, fmt.Errorf("project with id %d: %w", projectId, custom_errors.ErrProjectNotFound)
		}
		return models.Project{}, fmt.Errorf("get project %d: %w", projectId, err)
	}
	return project, nil
}

func (s *ProjectService) CreateProject(ctx context.Context, body *models.CreateProjectRequest) (models.Project, error) {
	if err := body.Validate(); err != nil {
		return models.Project{}, fmt.Errorf("create project: %w", custom_errors.ErrInvalidProject)
	}

	// creating project object and saving it to database
	project := models.Project{
		User:         uint32(*body.OwnerID),
		Title:        *body.Title,
		Dependencies: body.Dependencies,
	}

	if err := s.r.CreateProject(ctx, &project); err != nil {
		return project, fmt.Errorf("failed to create project: %w", err)
	}

	if err := s.createProjectDir(&project); err != nil {
		return project, fmt.Errorf("create project dir: %w", custom_errors.ErrServerError)
	}

	// updating project path
	if err := s.updateProjectPath(ctx, &project); err != nil {
		return project, fmt.Errorf("update project dir: %w", custom_errors.ErrServerError)
	}

	return project, nil
}

func (s *ProjectService) EditProject(ctx context.Context, projectId uint32, body *models.CreateProjectRequest) (models.Project, error) {
	project, err := s.r.GetProjectById(ctx, projectId)
	if err != nil {
		return project, err
	}

	project.Dependencies = body.Dependencies
	project.Title = *body.Title

	if err = s.r.UpdateProject(ctx, &project); err != nil {
		return project, err
	}
	return project, nil
}

// os.MkdirAll does not accept a context, so none is passed here
func (s *ProjectService) createProjectDir(project *models.Project) error {
	// fmt.Println(s.cfg.ProjectPath + strconv.FormatUint(uint64(project.ID), 10) + "/" + project.Title)
	if err := os.MkdirAll("./"+s.cfg.ProjectPath+strconv.FormatUint(uint64(project.ID), 10)+"/"+project.Title, 0755); err != nil {
		return fmt.Errorf("create project directory: %w", err)
	}
	return nil
}

func (s *ProjectService) updateProjectPath(ctx context.Context, project *models.Project) error {
	project.Path = strconv.FormatUint(uint64(project.ID), 10) + "/" + project.Title

	if err := s.r.UpdateProject(ctx, project); err != nil {
		return fmt.Errorf("update path to project directory: %w", custom_errors.ErrProjectNotFound)
	}
	return nil
}

// func (s *ProjectService) buildContainer() {
//
// }
