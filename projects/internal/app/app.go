package app

import (
	"context"
	"fmt"

	"github.com/go-pg/pg/v10"

	"workspaces/internal/platform/config"
	"workspaces/projects/internal/handler"
	"workspaces/projects/internal/repository"
	"workspaces/projects/internal/service"
)

type Handlers struct {
	Project *handler.ProjectHandler
	File    *handler.FileHandler
}

type Services struct {
	Project *service.ProjectService
	File    *service.FileService
}

type Repositories struct {
	Project *repository.ProjectRepository
	File    *repository.FileRepository
}

type App struct {
	Services     Services
	Handlers     Handlers
	Repositories Repositories
}

func New(ctx context.Context, cfg *config.Config, db *pg.DB) (*App, error) {
	err := createSchema(db)
	if err != nil {
		return nil, fmt.Errorf("create schema: %w", err)
	}

	repos := Repositories{
		Project: repository.NewProjectRepository(cfg, db),
		File:    repository.NewFileRepository(cfg, db),
	}
	svcs := Services{
		Project: service.NewProjectService(cfg, repos.Project),
		File:    service.NewFileService(cfg, repos.Project, repos.File),
	}
	hs := Handlers{
		Project: handler.NewProjectHandler(svcs.Project),
		File:    handler.NewFileHandler(svcs.File),
	}

	return &App{Repositories: repos, Services: svcs, Handlers: hs}, nil
}
