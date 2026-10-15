package service

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"path"
	"path/filepath"
	"workspaces/internal/platform/config"
	"workspaces/projects/internal/custom_errors"
	"workspaces/projects/internal/models"
	"workspaces/projects/internal/repository"
	"workspaces/projects/internal/util"
)

type FileService struct {
	project_r *repository.ProjectRepository
	file_r    *repository.FileRepository
	cfg       *config.Config
}

func NewFileService(config *config.Config, project_r *repository.ProjectRepository, file_r *repository.FileRepository) *FileService {
	return &FileService{
		project_r: project_r,
		file_r:    file_r,
		cfg:       config,
	}
}

// business logic
func (s *FileService) GetFilesList(ctx context.Context, project_id uint32) ([]models.File, error) {
	files, err := s.file_r.GetAllFilesByProjectID(ctx, project_id)
	if err != nil {
		return []models.File{}, fmt.Errorf("not found files with project_id: %d %w", project_id, custom_errors.ErrProjectNotFound)
	}

	return files, nil
}

func (s *FileService) CreateNewFile(ctx context.Context, body models.CreateFileRequest) (models.File, error) {
	if err := body.Validate(); err != nil {
		return models.File{}, err
	}

	// Имя от клиента: относительное, без "..".
	name, err := util.CleanRelPath(body.Filename)
	if err != nil {
		return models.File{}, fmt.Errorf("%w: %v", custom_errors.ErrInvalidFile, err)
	}

	project, err := s.project_r.GetProjectById(ctx, body.ProjectID)
	if err != nil {
		return models.File{}, fmt.Errorf("not found project with id: %d %w", body.ProjectID, custom_errors.ErrProjectNotFound)
	}

	projectPath, err := util.CleanRelPath(project.Path)
	if err != nil {
		return models.File{}, fmt.Errorf("%w: %v", custom_errors.ErrInvalidProject, err)
	}

	root, err := filepath.Abs(s.cfg.ProjectPath)
	if err != nil {
		return models.File{}, err
	}

	// Как и раньше: путь файла относительно PROJECT_PATH, например "1/test1/src/main.rs".
	rel := path.Join(projectPath, name)

	if err := util.CreateFileExclusive(root, rel); err != nil {
		if errors.Is(err, fs.ErrExist) {
			return models.File{}, fmt.Errorf("%w: %s", custom_errors.ErrFileExists, rel)
		}
		return models.File{}, fmt.Errorf("%w: %v", custom_errors.ErrInvalidFile, err)
	}

	file := models.File{ProjectID: project.ID, Path: name, Text: ""}
	if err := s.file_r.SaveFile(ctx, &file); err != nil {
		_ = util.RemoveFile(root, rel) // не оставляем осиротевший файл
		return models.File{}, fmt.Errorf("%w: %v", custom_errors.ErrInvalidFile, err)
	}

	return file, nil
}

func (s *FileService) GetFile(ctx context.Context, id uint32) (models.File, error) {
	file, err := s.file_r.GetFileByID(ctx, id)
	if err != nil {
		return models.File{}, fmt.Errorf("%w: %v", custom_errors.ErrInvalidFile, err)
	}

	// it is for checking rights in future
	// project, err := s.project_r.GetProjectById(ctx, file.ProjectID)
	// if err != nil {
	//  return models.File{}, fmt.Errorf("not found project with id: %d %w", file.ProjectID, custom_errors.ErrProjectNotFound)
	// }

	return file, nil
}

func (s *FileService) EditFile(ctx context.Context, body *models.EditFileRequest) (models.File, error) {
	file, err := s.file_r.GetFileByID(ctx, body.ID)
	if err != nil {
		return models.File{}, fmt.Errorf("%w: %v", custom_errors.ErrInvalidFile, err)
	}

	proj, err := s.project_r.GetProjectById(ctx, file.ProjectID)
	if err != nil {
		return models.File{}, fmt.Errorf("%w: %v", custom_errors.ErrInvalidFile, err)
	}

	// TODO: implement checking rights by a jwt token
	// file.Path уже относителен PROJECT_PATH, поэтому берём его как есть.
	rel, err := util.CleanRelPath(file.Path)
	if err != nil {
		return file, fmt.Errorf("%w: %v", custom_errors.ErrInvalidFile, err)
	}

	projectDir, err := util.ProjectDir(s.cfg.ProjectPath, proj.Path)
	if err != nil {
		return file, fmt.Errorf("%w: %v", custom_errors.ErrInvalidFile, err)
	}

	if err := util.WriteFileAtomic(projectDir, rel, []byte(body.Text)); err != nil {
		return file, fmt.Errorf("%w: %v", custom_errors.ErrInvalidFile, err)
	}

	file.Text = body.Text
	if err := s.file_r.UpdateFile(ctx, &file); err != nil {
		return file, fmt.Errorf("%w: %v", custom_errors.ErrInvalidFile, err)
	}

	return file, nil
}
