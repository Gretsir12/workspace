package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
	"workspaces/internal/platform/config"
	"workspaces/projects/internal/models"
	"workspaces/projects/internal/repository"
	"workspaces/projects/internal/util"
)

const (
	nixpkgsRev   = "6510408d147d8c5e6d76043e899e9251756f6a75"
	workspaceTTL = 8 * time.Hour

	// Параметры ожидания изменений на диске после команды.
	syncPollInterval = 300 * time.Millisecond
	syncMaxWait      = 60 * time.Second
	syncIdleTicks    = 10 // ~3 с без изменений: считаем, что команда ничего не писала
	syncStableTicks  = 2  // столько тиков подряд диск должен не меняться после изменения
)

var baseDependencies = []string{"gcc", "gnumake", "fzf", "tree", "pkg-config"}

type WorkspaceService struct {
	cfg         *config.Config
	project_r   *repository.ProjectRepository
	workspace_r *repository.WorkspaceRepository
	image_r     *repository.ImageRepository
	file_r      *repository.FileRepository
	docker      *util.Docker

	syncMu sync.Map // projectID (uint32) -> *sync.Mutex
}

func NewWorkspaceService(
	config *config.Config,
	project_r *repository.ProjectRepository,
	workspace_r *repository.WorkspaceRepository,
	image_r *repository.ImageRepository,
	file_r *repository.FileRepository,
	docker *util.Docker,
) *WorkspaceService {
	return &WorkspaceService{
		cfg:         config,
		project_r:   project_r,
		workspace_r: workspace_r,
		image_r:     image_r,
		file_r:      file_r,
		docker:      docker,
	}
}

// OpenWorkspace собирает образ (лог сборки пишется в logs), сохраняет образ
// и workspace в БД. Сборку запускаем всегда: при готовом образе она отработает
// из кеша слоёв и заодно гарантирует, что образ реально есть в Docker.
func (s *WorkspaceService) OpenWorkspace(ctx context.Context, body *models.CreateWorkspaceRequest, logs io.Writer) (models.Workspace, error) {

	deps := append(body.Dependencies, baseDependencies...)

	tag, err := s.docker.BuildImageWithDeps(ctx, deps, logs)
	if err != nil {
		return models.Workspace{}, err
	}

	image := models.Image{
		Hash:       util.HashDependencies(deps),
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
func (s *WorkspaceService) StartTerminal(ctx context.Context, w models.Workspace, projId uint32) (*util.Terminal, error) {
	project, err := s.project_r.GetProjectById(ctx, projId)
	if err != nil {
		return nil, err
	}

	term, err := s.docker.StartTerminal(ctx, w.Image.Image, w.ID, s.cfg.ProjectPath+"/"+project.Path)
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

// ForwardInput пропускает кадр от клиента в терминал. На Enter берёт снимок
// директории «до» и запускает в фоне ожидание изменений на диске.
// line хранит приблизительную копию набираемой строки между кадрами.
func (s *WorkspaceService) ForwardInput(
	ctx context.Context,
	term io.Writer,
	line *[]byte,
	data []byte,
	workspaceID uint32,
	projectID uint32,
) error {
	start := 0
	for i, b := range data {
		switch b {
		case '\r', '\n':
			cmd := strings.TrimSpace(string(*line))
			*line = (*line)[:0]

			if _, err := term.Write(data[start:i]); err != nil {
				return err
			}

			var (
				before     string
				haveBefore bool
			)
			if cmd != "" {
				snap, err := s.getSnapshot(ctx, projectID)
				if err != nil {
					log.Printf("before hook: ws=%d: %v", workspaceID, err)
				} else {
					before = snap
					haveBefore = true
				}
			}

			if _, err := term.Write(data[i : i+1]); err != nil {
				return err
			}

			// «После» берём не сразу: Write в PTY лишь кладёт байт в буфер,
			// команда к этому моменту ещё не выполнилась.
			if haveBefore {
				go s.syncAfterCommand(workspaceID, projectID, before)
			}

			start = i + 1

		case 0x7f, 0x08: // Backspace
			if len(*line) > 0 {
				_, size := utf8.DecodeLastRune(*line)
				*line = (*line)[:len(*line)-size]
			}
		case 0x03, 0x15: // Ctrl+C, Ctrl+U
			*line = (*line)[:0]
		default:
			if b >= 0x20 {
				*line = append(*line, b)
			}
		}
	}

	_, err := term.Write(data[start:])
	return err
}

// syncAfterCommand опрашивает диск, пока снимок не изменится и не устоится,
// после чего синхронизирует файлы проекта с БД. Работает в своей горутине
// со своим контекстом: контекст запроса к этому моменту может быть отменён.
func (s *WorkspaceService) syncAfterCommand(workspaceID, projectID uint32, before string) {
	ctx, cancel := context.WithTimeout(context.Background(), syncMaxWait)
	defer cancel()

	// Путь проекта не меняется, поэтому достаём из БД один раз, а не на каждый тик.
	project, err := s.project_r.GetProjectById(ctx, projectID)
	if err != nil {
		log.Printf("sync: ws=%d: get project: %v", workspaceID, err)
		return
	}
	root := s.cfg.ProjectPath + "/" + project.Path

	ticker := time.NewTicker(syncPollInterval)
	defer ticker.Stop()

	last := before
	stable := 0
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}

		snap, err := util.SnapshotDirectory(root)
		if err != nil {
			log.Printf("sync: ws=%d: snapshot: %v", workspaceID, err)
			continue
		}

		if snap != last {
			last = snap // диск ещё меняется, ждём тишины
			stable = 0
			continue
		}

		stable++
		if last != before && stable >= syncStableTicks {
			log.Printf("sync: ws=%d: project %d changed, syncing", workspaceID, projectID)
			s.syncProject(ctx, projectID, root)
			return
		}
		if last == before && stable >= syncIdleTicks {
			return // команда ничего не изменила
		}
	}
}

func (s *WorkspaceService) syncProject(ctx context.Context, projectID uint32, root string) {
	defer s.lockProject(projectID)()

	files, err := util.ScanDirectory(root, projectID)
	if err != nil {
		log.Printf("sync: project %d: scan: %v", projectID, err)
		return
	}
	if err := s.updateFiles(ctx, projectID, root, files); err != nil {
		log.Printf("sync: project %d: update files: %v", projectID, err)
	}
}

// lockProject не даёт двум синхронизациям одного проекта идти параллельно
// (иначе возможны дубли записей в БД). Возвращает функцию разблокировки.
func (s *WorkspaceService) lockProject(id uint32) func() {
	m, _ := s.syncMu.LoadOrStore(id, &sync.Mutex{})
	mu := m.(*sync.Mutex)
	mu.Lock()
	return mu.Unlock
}

// updateFiles добавляет в БД новые файлы и удаляет записи файлов,
// которых больше нет на диске.
func (s *WorkspaceService) updateFiles(ctx context.Context, projectID uint32, root string, files []models.File) error {
	log.Println("Updating files")

	existing, err := s.file_r.GetAllFilesByProjectID(ctx, projectID)
	if err != nil {
		return fmt.Errorf("load existing files: %w", err)
	}

	known := make(map[string]struct{}, len(existing))
	for _, f := range existing {
		known[f.Path] = struct{}{}
	}

	var errs []error
	for i := range files {
		f := &files[i]
		if _, ok := known[f.Path]; ok {
			continue
		}
		f.ProjectID = projectID
		if err := s.file_r.SaveFile(ctx, f); err != nil {
			errs = append(errs, fmt.Errorf("save %s: %w", f.Path, err))
			continue
		}
		known[f.Path] = struct{}{} // защита от дублей внутри самого files
	}

	// Удаляем записи, которых нет на диске. Сначала смотрим в результат скана,
	// а для оставшихся проверяем диск напрямую: так не удалим файл, который
	// просто не попал в скан (игнор-лист, размер, не UTF-8). Заодно уйдут
	// старые записи с абсолютными путями: по ним файла в root нет.
	scanned := make(map[string]struct{}, len(files))
	for _, f := range files {
		scanned[f.Path] = struct{}{}
	}
	for _, f := range existing {
		if _, ok := scanned[f.Path]; ok {
			continue
		}
		_, statErr := os.Stat(filepath.Join(root, filepath.FromSlash(f.Path)))
		if !errors.Is(statErr, os.ErrNotExist) {
			continue // файл на месте, либо не смогли проверить: не трогаем
		}
		if err := s.file_r.DeleteFile(ctx, f.ID); err != nil {
			errs = append(errs, fmt.Errorf("delete %s: %w", f.Path, err))
		}
	}
	return errors.Join(errs...)
}

func (s *WorkspaceService) getSnapshot(ctx context.Context, projectID uint32) (string, error) {
	project, err := s.project_r.GetProjectById(ctx, projectID)
	if err != nil {
		return "", err
	}
	return util.SnapshotDirectory(s.cfg.ProjectPath + "/" + project.Path)
}
