package repository

import (
	"workspaces/internal/platform/config"
	"workspaces/projects/internal/models"

	"github.com/go-pg/pg/v10"
)

type FileRepository struct {
	cfg *config.Config
	db  *pg.DB
}

func NewFileRepository(config *config.Config, db *pg.DB) *FileRepository {
	return &FileRepository{
		cfg: config,
		db:  db,
	}
}

func (r *FileRepository) SaveFile(file *models.File) error {
	_, err := r.db.Model(file).Insert()
	return err
}

func (r *FileRepository) GetFileByID(id uint32) (models.File, error) {
	file := models.File{ID: id}
	if err := r.db.Model(&file).WherePK().Select(); err != nil {
		return file, err
	}
	return file, nil
}

func (r *FileRepository) GetFileByPath(path string) (models.File, error) {
	file := models.File{}
	err := r.db.Model(&file).
		Where("path = ?", path).
		Select()
	if err != nil {
		return models.File{}, err
	}
	return file, nil
}

func (r *FileRepository) UpdateFile(file *models.File) error {
	_, err := r.db.Model(file).WherePK().Update()
	if err != nil {
		return err
	}
	return nil
}

func (r *FileRepository) GetAllFilesByProjectID(project_id uint32) ([]models.File, error) {
	files := []models.File{}
	err := r.db.Model(&files).
		Where("project_id = ?", project_id).
		Select()
	if err != nil {
		return []models.File{}, err
	}
	return files, nil
}
