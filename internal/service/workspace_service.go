package service

import (
	"workspaces/projects/internal/repository"
	"workspaces/projects/internal/util"
)

type WorkspaceService struct {
	cfg         *util.Config
	project_r   repository.ProjectRepository
	workspace_r repository.WorkspaceRepository
}

func NewWorkspaceService(config *util.Config, project_r *repository.ProjectRepository, workspace_r repository.WorkspaceRepository) *WorkspaceService {
	return &WorkspaceService{
		cfg:         config,
		project_r:   *project_r,
		workspace_r: workspace_r,
	}
}

func (s *WorkspaceService) CreateNewWorkspace() {}
