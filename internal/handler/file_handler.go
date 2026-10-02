package handler

import (
	"encoding/json"
	"log"
	"strconv"
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
	r.Get("/{id}", h.Get)
	r.Put("/", h.Update)

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
	id, err := strconv.ParseUint(r.PathValue("id"), 10, 32)
	if err != nil {
		http.Error(w, "invalid project_id", custom_errors.StatusFromError(err))
		return
	}

	file, err := h.s.GetFile(uint32(id))
	if err != nil {
		http.Error(w, err.Error(), custom_errors.StatusFromError(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(file); err != nil {
		log.Printf("encode project %d: %v", file.Path, err)
	}
}

func (h *FileHandler) Update(w http.ResponseWriter, r *http.Request) {
	var b models.EditFileRequest

	err := json.NewDecoder(r.Body).Decode(&b)
	if err != nil {
		http.Error(w, err.Error(), custom_errors.StatusFromError(err))
		return
	}

	file, err := h.s.EditFile(&b)
	if err != nil {
		http.Error(w, err.Error(), custom_errors.StatusFromError(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(file); err != nil {
		log.Printf("encode project %d: %v", file.Path, err)
	}
}
