package service

import (
	"fmt"
	"os"
	"errors"
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

func getFileText(path string) (string, error) {
	dat, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	fmt.Println(string(dat))
	
	return string(dat), nil
}
func checkFileExists(path string) bool {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
	  return false
	}
	return true
}

func (s *FileService) CreateNewFile(body models.CreateFileRequest) (models.File, error) {
	if err := body.Validate(); err != nil {
		return models.File{}, err
	}

	project, err := s.project_r.GetProjectById(body.ProjectID)
	if err != nil {
		return models.File{}, fmt.Errorf("not found project with id: %d %w", body.ProjectID, custom_errors.ErrProjectNotFound)
	}

	if checkFileExists(project.Path+"/"+body.Filename) {
		return models.File{}, fmt.Errorf("%w: %v", custom_errors.ErrFileExists, project.Path+"/"+body.Filename)
	}

	if err := writeFile(project.Path+"/"+body.Filename, ""); err != nil {
		fmt.Println("writeFile failed", "path", project.Path+"/"+body.Filename, "err", err)
		return models.File{}, fmt.Errorf("%w: %v", custom_errors.ErrInvalidFile, err)
	}

	file := models.File{ProjectID: project.ID, Path: project.Path + "/" + body.Filename, Text: ""}
	if err = s.file_r.SaveFile(&file); err != nil {
		return models.File{}, fmt.Errorf("%w: %v", custom_errors.ErrInvalidFile, err)
	}

	return file, nil
}

func (s *FileService) GetFile(id uint32) (models.File, error) {

	file, err := s.file_r.GetFileByID(id)
	if err != nil {
		return models.File{}, fmt.Errorf("%w: %v", custom_errors.ErrInvalidFile, err)
	}

	// it is for checking rights in future
	// project, err := s.project_r.GetProjectById(file.ProjectID)
	// if err != nil {
	// 	return models.File{}, fmt.Errorf("not found project with id: %d %w", file.ProjectID, custom_errors.ErrProjectNotFound)
	// }

	return file, nil
}

func (s *FileService) EditFile(body *models.EditFileRequest) (models.File, error) {
	file, err := s.file_r.GetFileByID(body.ID)
	if err != nil {
		return models.File{}, fmt.Errorf("%w: %v", custom_errors.ErrInvalidFile, err)
	}

	// TODO: implement checking rights by a jwt token

	writeFile(file.Path, body.Text)
	file.Text, err = getFileText(file.Path)	
	if err != nil {
		return file, fmt.Errorf("%w: %v", custom_errors.ErrInvalidFile, err)
	}

	if err = s.file_r.UpdateFile(&file); err != nil {
		return file, fmt.Errorf("%w: %v", custom_errors.ErrInvalidFile, err)
	}

	return file, nil
}
