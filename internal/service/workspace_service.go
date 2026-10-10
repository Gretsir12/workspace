package service

import (
	"workspaces/internal/platform/config"
	"workspaces/projects/internal/repository"
)

type WorkspaceService struct {
	cfg         *config.Config
	project_r   repository.ProjectRepository
	workspace_r repository.WorkspaceRepository
}

func NewWorkspaceService(config *config.Config, project_r *repository.ProjectRepository, workspace_r repository.WorkspaceRepository) *WorkspaceService {
	return &WorkspaceService{
		cfg:         config,
		project_r:   *project_r,
		workspace_r: workspace_r,
	}
}

func (s *WorkspaceService) CreateNewWorkspace() {

}
