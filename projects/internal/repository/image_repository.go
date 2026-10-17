package repository

import (
	"context"
	"workspaces/internal/platform/config"
	"workspaces/projects/internal/models"

	"github.com/go-pg/pg/v10"
)

type ImageRepository struct {
	cfg *config.Config
	db  *pg.DB
}

func NewWorkspaceImageRepository(cfg *config.Config, db *pg.DB) *ImageRepository {
	return &ImageRepository{
		cfg: cfg,
		db:  db,
	}
}

// UpsertImage сохраняет образ; при повторном использовании обновляет
// тег, статус и last_used_at (по last_used_at можно чистить старые образы).
func (r *ImageRepository) UpsertImage(ctx context.Context, image *models.Image) error {
	_, err := r.db.ModelContext(ctx, image).
		OnConflict("(hash) DO UPDATE").
		Set("image = EXCLUDED.image, status = EXCLUDED.status, updated_at = now(), last_used_at = now()").
		Insert()
	return err
}

// FindImageByHash возвращает pg.ErrNoRows, если образа нет.
func (r *ImageRepository) FindImageByHash(ctx context.Context, hash string) (*models.Image, error) {
	image := &models.Image{}
	if err := r.db.ModelContext(ctx, image).Where("hash = ?", hash).Select(); err != nil {
		return nil, err
	}
	return image, nil
}
