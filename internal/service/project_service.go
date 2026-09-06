package service

import "workspaces/projects/internal/repository"

type ProjectService struct {
	r *repository.ProjectRepository
}

func NewProjectService(r *repository.ProjectRepository) *ProjectService {
	return &ProjectService{
		r: r,
	}
}
