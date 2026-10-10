package projects

import (
	"context"
	"fmt"

	"github.com/go-chi/chi/v5"
	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"

	"workspaces/internal/platform/config"
	"workspaces/projects/internal/handler"
	"workspaces/projects/internal/models"
	"workspaces/projects/internal/repository"
	"workspaces/projects/internal/service"
)

type Module struct {
	project *handler.ProjectHandler
	file    *handler.FileHandler
}

func New(ctx context.Context, cfg *config.Config, db *pg.DB) (*Module, error) {
	if err := createSchema(db); err != nil {
		return nil, fmt.Errorf("create schema: %w", err)
	}

	projectRepo := repository.NewProjectRepository(cfg, db)
	fileRepo := repository.NewFileRepository(cfg, db)

	projectSvc := service.NewProjectService(cfg, projectRepo)
	fileSvc := service.NewFileService(cfg, projectRepo, fileRepo)

	return &Module{
		project: handler.NewProjectHandler(projectSvc),
		file:    handler.NewFileHandler(fileSvc),
	}, nil
}

func (m *Module) Routes(r chi.Router) {
	r.Get("/proj/{id}", m.project.Get)
	r.Post("/proj", m.project.Create)

	r.Get("/files", m.file.Get)
	r.Get("/files/{project_id}", m.file.GetAllProjectFiles)
	r.Post("/files", m.file.Create)
	r.Put("/files", m.file.Update)
}

func createSchema(db *pg.DB) error {
	models := []interface{}{
		(*models.Project)(nil),
		(*models.File)(nil),
		(*models.Workspace)(nil),
	}

	for _, model := range models {
		err := db.Model(model).CreateTable(&orm.CreateTableOptions{
			Temp:          false,
			FKConstraints: true,
			IfNotExists:   true,
		})
		if err != nil {
			return err
		}
	}
	return nil
}
