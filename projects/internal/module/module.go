package module

import (
	"log"
	platform "workspaces/internal/platform/module"
	"workspaces/projects/internal/handler"
	"workspaces/projects/internal/repository"
	"workspaces/projects/internal/service"
	"workspaces/projects/internal/util"
)

type Handlers struct {
	Project   *handler.ProjectHandler
	File      *handler.FileHandler
	Workspace *handler.WorkspaceHandler
}

type Module struct {
	Handlers *Handlers
}

const (
	nixpkgsRev = "6510408d147d8c5e6d76043e899e9251756f6a75"
)

// проверка на этапе компиляции, что Module реализует интерфейс
var _ platform.Module = (*Module)(nil)

func New(deps platform.Deps) *Module {

	createSchema(deps.DB)

	dockerUtil, err := util.NewDocker(nixpkgsRev)
	if err != nil {
		log.Fatal("Failed to start docker client: %v", err)
	}

	projectRepo := repository.NewProjectRepository(deps.Cfg, deps.DB)
	fileRepo := repository.NewFileRepository(deps.Cfg, deps.DB)
	workspaceRepo := repository.NewWorkspaceRepository(deps.Cfg, deps.DB)
	imageRepo := repository.NewWorkspaceImageRepository(deps.Cfg, deps.DB)

	projectSvc := service.NewProjectService(deps.Cfg, projectRepo)
	fileSvc := service.NewFileService(deps.Cfg, projectRepo, fileRepo)
	workspaceSvc := service.NewWorkspaceService(deps.Cfg, projectRepo, workspaceRepo, imageRepo, dockerUtil)

	return &Module{
		Handlers: &Handlers{
			Project:   handler.NewProjectHandler(projectSvc),
			File:      handler.NewFileHandler(fileSvc),
			Workspace: handler.NewWorkspaceHandler(workspaceSvc),
		},
	}
}

func (m *Module) Name() string { return "projects" }
