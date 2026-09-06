package handler

import "workspaces/projects/internal/service"

type ProjectHandler struct {
	s *service.ProjectService
}

func NewProjectHandler(s *service.ProjectService) *ProjectHandler {
	return &ProjectHandler{
		s: s,
	}
}
