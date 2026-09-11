package repository

import "workspaces/projects/internal/util"

type ProjectRepository struct {
	cfg *util.Config
}

func NewProjectRepository(config *util.Config) *ProjectRepository {
	return &ProjectRepository{
		cfg: config,
	}
}
