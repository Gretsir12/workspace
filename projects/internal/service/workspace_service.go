package service

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"
	"workspaces/internal/platform/config"
	"workspaces/projects/internal/models"
	"workspaces/projects/internal/repository"
	"workspaces/projects/internal/util"
)

const (
	nixpkgsRev   = "6510408d147d8c5e6d76043e899e9251756f6a75"
	workspaceTTL = 8 * time.Hour // TODO: вынести в конфиг
)

type WorkspaceService struct {
	cfg         *config.Config
	project_r   *repository.ProjectRepository
	workspace_r *repository.WorkspaceRepository
	image_r     *repository.ImageRepository
	docker      *util.Docker
}

func NewWorkspaceService(
	config *config.Config,
	project_r *repository.ProjectRepository,
	workspace_r *repository.WorkspaceRepository,
	image_r *repository.ImageRepository,
	docker *util.Docker,
) *WorkspaceService {
	return &WorkspaceService{
		cfg:         config,
		project_r:   project_r,
		workspace_r: workspace_r,
		image_r:     image_r,
		docker:      docker,
	}
}

// OpenWorkspace собирает образ (лог сборки пишется в logs), сохраняет образ
// и workspace в БД. Сборку запускаем всегда: при готовом образе она отработает
// из кеша слоёв и заодно гарантирует, что образ реально есть в Docker.
func (s *WorkspaceService) OpenWorkspace(ctx context.Context, body *models.CreateWorkspaceRequest, logs io.Writer) (models.Workspace, error) {
	tag, err := s.docker.BuildImageWithDeps(ctx, body.Dependencies, logs)
	if err != nil {
		return models.Workspace{}, err
	}

	image := models.Image{
		Hash:       util.HashDependencies(body.Dependencies),
		NixpkgsRev: nixpkgsRev,
		Deps:       body.Dependencies,
		Image:      tag,
		Status:     models.ImageReady,
	}
	if err := s.image_r.UpsertImage(ctx, &image); err != nil {
		return models.Workspace{}, fmt.Errorf("save image: %w", err)
	}

	workspace := models.Workspace{
		UserID:    body.UserId,
		ImageHash: image.Hash,
		Image:     &image,
		Status:    models.WorkspacePending,
		ExpiresAt: time.Now().Add(workspaceTTL),
	}
	if err := s.workspace_r.CreateWorkspace(ctx, &workspace); err != nil {
		return models.Workspace{}, fmt.Errorf("save workspace: %w", err)
	}
	return workspace, nil
}

// StartTerminal запускает контейнер и записывает его ID и статус в БД.
func (s *WorkspaceService) StartTerminal(ctx context.Context, w models.Workspace) (*util.Terminal, error) {
	term, err := s.docker.StartTerminal(ctx, w.Image.Image, w.ID)
	if err != nil {
		s.setStatus(w.ID, models.WorkspaceFailed)
		return nil, err
	}
	if err := s.workspace_r.SetContainer(ctx, w.ID, term.ID(), models.WorkspaceRunning); err != nil {
		term.Close()
		s.setStatus(w.ID, models.WorkspaceFailed)
		return nil, fmt.Errorf("save container id: %w", err)
	}
	return term, nil
}

// StopWorkspace вызывается при закрытии сессии.
func (s *WorkspaceService) StopWorkspace(id uint32) {
	s.setStatus(id, models.WorkspaceStopped)
}

// Свой контекст: к моменту закрытия сессии контекст запроса уже отменён,
// и UPDATE с ним просто не выполнился бы.
func (s *WorkspaceService) setStatus(id uint32, status string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.workspace_r.UpdateStatus(ctx, id, status); err != nil {
		log.Printf("update workspace %d status %q: %v", id, status, err)
	}
}
