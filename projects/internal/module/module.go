// projects/internal/module/module.go
package module

import (
	platform "workspaces/internal/platform/module"
	"workspaces/projects/internal/handler"
	"workspaces/projects/internal/repository"
	"workspaces/projects/internal/service"
)

type Handlers struct {
	Project *handler.ProjectHandler
	File    *handler.FileHandler
}

type Module struct {
	Handlers *Handlers
}

// проверка на этапе компиляции, что Module реализует интерфейс
var _ platform.Module = (*Module)(nil)

func New(deps platform.Deps) *Module {
	projectRepo := repository.NewProjectRepository(deps.Cfg, deps.DB)
	fileRepo := repository.NewFileRepository(deps.Cfg, deps.DB)

	projectSvc := service.NewProjectService(deps.Cfg, projectRepo)
	fileSvc := service.NewFileService(deps.Cfg, projectRepo, fileRepo)

	return &Module{
		Handlers: &Handlers{
			Project: handler.NewProjectHandler(projectSvc),
			File:    handler.NewFileHandler(fileSvc),
		},
	}
}

func (m *Module) Name() string { return "projects" }
