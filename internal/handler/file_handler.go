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
	"workspaces/projects/internal/util"
)

type FileHandler struct {
	s *service.FileService
}

func NewFileHandler(service *service.FileService) *FileHandler {
	return &FileHandler{
		s: service,
	}
}

func (h *FileHandler) Create(w http.ResponseWriter, r *http.Request) {
	var b models.CreateFileRequest

	err := json.NewDecoder(r.Body).Decode(&b)
	if err != nil {
		httpserver.ResponseError(w, fmt.Sprintf("Failed to create file: %v", err), http.StatusBadRequest)
		return
	}

	file, err := h.s.CreateNewFile(r.Context(), b)
	if err != nil {
		httpserver.ResponseError(w, fmt.Sprintf("Failed to create file: %v", err), custom_errors.StatusFromError(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(file); err != nil {
		httpserver.ResponseError(w, fmt.Sprintf("Failed to create file: %v", err), custom_errors.StatusFromError(err))
	}
}

func (h *FileHandler) Get(w http.ResponseWriter, r *http.Request) {
	fileID, err := util.ParseUintFromQuery("file_id", r)
	if err != nil {
		httpserver.ResponseError(w, fmt.Sprintf("Failed to get file: %v", err), http.StatusBadRequest)
		return
	}

	file, err := h.s.GetFile(r.Context(), uint32(fileID))
	if err != nil {
		httpserver.ResponseError(w, fmt.Sprintf("Failed to get file: %v", err), custom_errors.StatusFromError(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(file); err != nil {
		httpserver.ResponseError(w, fmt.Sprintf("Failed to get file: %v", err), custom_errors.StatusFromError(err))
	}
}

func (h *FileHandler) Update(w http.ResponseWriter, r *http.Request) {
	var b models.EditFileRequest

	err := json.NewDecoder(r.Body).Decode(&b)
	if err != nil {
		httpserver.ResponseError(w, fmt.Sprintf("Failed to update file: %v", err), http.StatusBadRequest)
		return
	}

	file, err := h.s.EditFile(r.Context(), &b)
	if err != nil {
		httpserver.ResponseError(w, fmt.Sprintf("Failed to update file: %v", err), custom_errors.StatusFromError(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(file); err != nil {
		httpserver.ResponseError(w, fmt.Sprintf("Failed to update file: %v", err), custom_errors.StatusFromError(err))
		return
	}
}

func (h *FileHandler) GetAllProjectFiles(w http.ResponseWriter, r *http.Request) {
	project_id, err := strconv.ParseUint(r.PathValue("project_id"), 10, 32)
	if err != nil {
		httpserver.ResponseError(w, fmt.Sprintf("Failed to get files: %v", err), http.StatusBadRequest)
		return
	}

	files, err := h.s.GetFilesList(r.Context(), uint32(project_id))
	if err != nil {
		httpserver.ResponseError(w, fmt.Sprintf("Failed to get files: %v", err), custom_errors.StatusFromError(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(files); err != nil {
		httpserver.ResponseError(w, fmt.Sprintf("Failed to get files: %v", err), custom_errors.StatusFromError(err))
		return
	}
}
