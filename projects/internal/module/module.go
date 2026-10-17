package module

import (
	"context"
	"log"
	"time"
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
	docker   *util.Docker
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
		log.Fatalf("Failed to start docker client: %v", err)
	}

	projectRepo := repository.NewProjectRepository(deps.Cfg, deps.DB)
	fileRepo := repository.NewFileRepository(deps.Cfg, deps.DB)
	workspaceRepo := repository.NewWorkspaceRepository(deps.Cfg, deps.DB)
	imageRepo := repository.NewWorkspaceImageRepository(deps.Cfg, deps.DB)

	projectSvc := service.NewProjectService(deps.Cfg, projectRepo)
	fileSvc := service.NewFileService(deps.Cfg, projectRepo, fileRepo)
	workspaceSvc := service.NewWorkspaceService(deps.Cfg, projectRepo, workspaceRepo, imageRepo, fileRepo, dockerUtil)

	return &Module{
		Handlers: &Handlers{
			Project:   handler.NewProjectHandler(projectSvc),
			File:      handler.NewFileHandler(fileSvc),
			Workspace: handler.NewWorkspaceHandler(workspaceSvc),
		},
		docker: dockerUtil,
	}
}

func (m *Module) Shutdown(ctx context.Context) error {
	err := m.Handlers.Workspace.Shutdown(ctx)

	// Всегда добиваем оставшиеся контейнеры: в том числе Created,
	// которые не принадлежат ни одной сессии.
	fctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if n, ferr := m.docker.RemoveManagedContainers(fctx); ferr != nil {
		log.Printf("force cleanup: %v", ferr)
	} else if n > 0 {
		log.Printf("force cleanup: removed %d leftover containers", n)
	}

	_ = m.docker.Close()
	return err
}

func (m *Module) Name() string { return "projects" }
