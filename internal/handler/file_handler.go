package handler

import (
	"encoding/json"
	"log"
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
	r.Get("/", h.Get)

	return r
}

func (h *FileHandler) Create(w http.ResponseWriter, r *http.Request) {
	var b models.CreateFileRequest

	err := json.NewDecoder(r.Body).Decode(&b)
	if err != nil {
		http.Error(w, err.Error(), custom_errors.StatusFromError(err))
		return
	}

	file, err := h.s.CreateNewFile(b)
	if err != nil {
		http.Error(w, err.Error(), custom_errors.StatusFromError(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(file); err != nil {
		log.Printf("encode project %d: %v", file.Path, err)
	}
}

func (h *FileHandler) Get(w http.ResponseWriter, r *http.Request) {
	var b models.GetFileRequest	

	err := json.NewDecoder(r.Body).Decode(&b)
	if err != nil {
		http.Error(w, err.Error(), custom_errors.StatusFromError(err))
		return
	}

	file, err := h.s.GetFile(b)
	if err != nil {
		http.Error(w, err.Error(), custom_errors.StatusFromError(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(file); err != nil {
		log.Printf("encode project %d: %v", file.Path, err)
	}
}
