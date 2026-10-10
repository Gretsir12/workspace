package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"workspaces/projects/internal/custom_errors"
	"workspaces/projects/internal/models"
	"workspaces/projects/internal/service"
	"workspaces/projects/internal/util"

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
	// r.Get("/{id}", h.Get)
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
		log.Printf("encode project %s: %v", file.Path, err)
	}
}

func (h *FileHandler) Get(w http.ResponseWriter, r *http.Request) {
	fileID, err := util.ParseUintFromQuery("file_id", r)
	if err != nil {
		http.Error(w, "file_id query parameter is missing or invalid", http.StatusBadRequest)
	}

	file, err := h.s.GetFile(uint32(fileID))
	if err != nil {
		http.Error(w, err.Error(), custom_errors.StatusFromError(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(file); err != nil {
		log.Printf("encode file %s: %v", file.Path, err)
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

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(file); err != nil {
		log.Printf("encode project %s: %v", file.Path, err)
	}
}

func (h *FileHandler) GetAllProjectFiles(w http.ResponseWriter, r *http.Request) {
	project_id, err := strconv.ParseUint(r.PathValue("project_id"), 10, 32)
	if err != nil {
		http.Error(w, "invalid project_id", http.StatusBadRequest)
		return
	}

	files, err := h.s.GetFilesList(uint32(project_id))
	if err != nil {
		http.Error(w, err.Error(), custom_errors.StatusFromError(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(files); err != nil {
		log.Printf("encode files error: %v", err)
	}
}
