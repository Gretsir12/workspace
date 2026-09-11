package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"workspaces/projects/internal/models"
	"workspaces/projects/internal/service"

	"github.com/go-chi/chi/v5"
)

type ProjectHandler struct {
	s *service.ProjectService
}

func NewProjectHandler(service *service.ProjectService) *ProjectHandler {
	return &ProjectHandler{
		s: service,
	}
}

func (h *ProjectHandler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/{id}", h.Get)
	r.Post("/", h.Create)

	return r
}

func (h *ProjectHandler) Get(w http.ResponseWriter, r *http.Request) {
	// ...
}

func (h *ProjectHandler) Create(w http.ResponseWriter, r *http.Request) {
	var b models.CreateProjectRequest

	err := json.NewDecoder(r.Body).Decode(&b)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = h.s.CreateProject(&b)
	if err != nil {
		http.Error(w, err.Error(), statusFromError(err))
	}

}

func statusFromError(err error) int {
	switch {
	case errors.Is(err, service.ErrProjectNotFound):
		return http.StatusNotFound

	case errors.Is(err, service.ErrProjectAlreadyExists):
		return http.StatusConflict

	case errors.Is(err, service.ErrInvalidProject):
		return http.StatusBadRequest

	default:
		return http.StatusInternalServerError
	}
}
