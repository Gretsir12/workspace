package repository

import "workspaces/projects/api"

type ProjectRepository struct {
	config *api.Config
}

func NewProjectRepository(c *api.Config) *ProjectRepository {
	return &ProjectRepository{
		config: c,
	}
}
