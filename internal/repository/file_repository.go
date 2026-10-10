package repository

import (
	"context"
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

func (r *FileRepository) SaveFile(ctx context.Context, file *models.File) error {
	_, err := r.db.ModelContext(ctx, file).Insert()
	return err
}

func (r *FileRepository) GetFileByID(ctx context.Context, id uint32) (models.File, error) {
	file := models.File{ID: id}
	if err := r.db.ModelContext(ctx, &file).WherePK().Select(); err != nil {
		return file, err
	}
	return file, nil
}

func (r *FileRepository) GetFileByPath(ctx context.Context, path string) (models.File, error) {
	file := models.File{}
	err := r.db.ModelContext(ctx, &file).
		Where("path = ?", path).
		Select()
	if err != nil {
		return models.File{}, err
	}
	return file, nil
}

func (r *FileRepository) UpdateFile(ctx context.Context, file *models.File) error {
	_, err := r.db.ModelContext(ctx, file).WherePK().Update()
	if err != nil {
		return err
	}
	return nil
}

func (r *FileRepository) GetAllFilesByProjectID(ctx context.Context, project_id uint32) ([]models.File, error) {
	files := []models.File{}
	err := r.db.ModelContext(ctx, &files).
		Where("project_id = ?", project_id).
		Select()
	if err != nil {
		return []models.File{}, err
	}
	return files, nil
}
