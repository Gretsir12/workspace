package handler

import (
	"encoding/json"
	"net/http"
	"workspaces/projects/internal/custom_errors"
	"workspaces/projects/internal/models"
	"workspaces/projects/internal/service"

	"github.com/go-chi/chi/v5"
)

type FileHandler struct {
	s *service.FileService
}

func NewFileHandler(service *service.FileService) *FileHandler {
	return &FileHandler{
		s: service,
	}
}

func (h *FileHandler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Post("/", h.Create)

	return r
}

func (h *FileHandler) Create(w http.ResponseWriter, r *http.Request) {
	var b models.CreateFileRequest

	err := json.NewDecoder(r.Body).Decode(&b)
	if err != nil {
		http.Error(w, err.Error(), custom_errors.StatusFromError(err))
		return
	}

	_, err = h.s.CreateNewFile(b)
	if err != nil {
		http.Error(w, err.Error(), custom_errors.StatusFromError(err))
		return
	}
}
