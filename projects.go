package projects

import (
	"context"

	"github.com/go-pg/pg/v10"

	"workspaces/internal/platform/config"
	"workspaces/projects/internal/handler"
	"workspaces/projects/internal/module"
	"workspaces/projects/internal/repository"
	"workspaces/projects/internal/service"
)

func New(ctx context.Context, cfg *config.Config, db *pg.DB) *module.Module {

	projectRepoisory := repository.NewProjectRepository(cfg, db)
	fileRepository := repository.NewFileRepository(cfg, db)

	projectService := service.NewProjectService(cfg, projectRepoisory)
	fileService := service.NewFileService(cfg, projectRepoisory, fileRepository)

	projectHandler := handler.NewProjectHandler(projectService)
	fileHandler := handler.NewFileHandler(fileService)

	return &module.Module{
		Ctx: ctx,
		Cfg: cfg,
		Db:  db,

		Handlers: &module.Handlers{
			Project: projectHandler,
			File:    fileHandler,
		},
	}
}
