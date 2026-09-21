package service

import (
	"fmt"
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

func (s *FileService) CreateNewFile(body models.CreateFileRequest) (models.File, error) {

	project, err := s.project_r.GetProjectById(body.ProjectID)
	if err != nil {
		return models.File{}, fmt.Errorf("create project directory: %w", err)
	}

	fmt.Println(project)

	// os.OpenFile(string(TMP_PATH))

	return models.File{}, nil // tmp
}
