package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"workspaces/internal/platform/httpserver"
	"workspaces/projects/internal/custom_errors"
	"workspaces/projects/internal/models"
	"workspaces/projects/internal/service"
)

type ProjectHandler struct {
	s *service.ProjectService
}

func NewProjectHandler(service *service.ProjectService) *ProjectHandler {
	return &ProjectHandler{
		s: service,
	}
}

func (h *ProjectHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(r.PathValue("id"), 10, 32)
	if err != nil {
		httpserver.ResponseError(w, fmt.Sprintf("Failed to get project: %v", err), custom_errors.StatusFromError(err))
		return
	}

	project, err := h.s.GetProject(uint32(id))
	if err != nil {
		httpserver.ResponseError(w, fmt.Sprintf("Failed to get project: %v", err), custom_errors.StatusFromError(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(project); err != nil {
		httpserver.ResponseError(w, fmt.Sprintf("Failed to get project: %v", err), custom_errors.StatusFromError(err))
	}
}

func (h *ProjectHandler) Create(w http.ResponseWriter, r *http.Request) {
	var b models.CreateProjectRequest

	err := json.NewDecoder(r.Body).Decode(&b)
	if err != nil {
		httpserver.ResponseError(w, fmt.Sprintf("Failed to create project: %v", err), http.StatusBadRequest)
		return
	}

	project, err := h.s.CreateProject(&b)
	if err != nil {
		httpserver.ResponseError(w, fmt.Sprintf("Failed to create project: %v", err), custom_errors.StatusFromError(err))
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(project); err != nil {
		httpserver.ResponseError(w, fmt.Sprintf("Failed to create project: %v", err), custom_errors.StatusFromError(err))
	}
}
