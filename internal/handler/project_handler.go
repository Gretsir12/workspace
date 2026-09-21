package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"workspaces/projects/internal/models"
	"workspaces/projects/internal/service"
	"workspaces/projects/internal/util"

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

	project, err := h.s.CreateProject(&b)
	if err != nil {
		http.Error(w, err.Error(), util.StatusFromError(err))
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // optional; 200 is the default
	if err := json.NewEncoder(w).Encode(project); err != nil {
		log.Printf("encode error: %v", err)
	}
}
