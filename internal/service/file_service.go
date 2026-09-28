package service

import (
	"fmt"
	"os"
	"path/filepath"
	"workspaces/projects/internal/custom_errors"
	"workspaces/projects/internal/models"
	"workspaces/projects/internal/repository"
	"workspaces/projects/internal/util"
)

type BASE_PATH string

var TMP_PATH BASE_PATH = "./tmp/"

type FileService struct {
	project_r *repository.ProjectRepository
	file_r    *repository.FileRepository
	cfg       *util.Config
}

func NewFileService(config *util.Config, project_r *repository.ProjectRepository, file_r *repository.FileRepository) *FileService {
	return &FileService{
		project_r: project_r,
		file_r:    file_r,
		cfg:       config,
	}
}

func writeFile(path string, text string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("mkdir %s: %w", dir, err)
	}
	return os.WriteFile(path, []byte(text), 0644)
}

func (s *FileService) CreateNewFile(body models.CreateFileRequest) (models.File, error) {
	project, err := s.project_r.GetProjectById(body.ProjectID)
	if err != nil {
		return models.File{}, fmt.Errorf("not found project with id: %d %w", body.ProjectID, custom_errors.ErrProjectNotFound)
	}

	if err := writeFile(project.Path+"/"+body.Filename, ""); err != nil {
		fmt.Println("writeFile failed", "path", project.Path+"/"+body.Filename, "err", err)
		return models.File{}, fmt.Errorf("%w: %v", custom_errors.ErrInvalidfile, err)
	}

	file := models.File{ProjectID: project.ID, Path: project.Path + "/" + body.Filename}
	if err = s.file_r.SaveFile(&file); err != nil {
		return models.File{}, fmt.Errorf("%w: %v", custom_errors.ErrInvalidfile, err)
	}

	return file, nil
}

// TODO: implemnt get file logic
func GetFileText(body models.GetFileRequest) (models.File, error) {
	return models.File{}, nil
}
