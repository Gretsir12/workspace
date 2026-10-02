package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"workspaces/projects/internal/custom_errors"
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
	id, err := strconv.ParseUint(r.PathValue("id"), 10, 32)
	if err != nil {
		http.Error(w, "invalid project_id", custom_errors.StatusFromError(err))
		return
	}

	project, err := h.s.GetProject(uint32(id))
	if err != nil {
		http.Error(w, "project is not exist or you cannot get it", custom_errors.StatusFromError(err))
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(project); err != nil {
		log.Printf("encode project %d: %v", id, err)
	}
}

func (h *ProjectHandler) Create(w http.ResponseWriter, r *http.Request) {
	var b models.CreateProjectRequest

	err := json.NewDecoder(r.Body).Decode(&b)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	project, err := h.s.CreateProject(&b)
	if err != nil {
		http.Error(w, err.Error(), custom_errors.StatusFromError(err))
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(project); err != nil {
		log.Printf("encode error: %v", err)
	}
}
