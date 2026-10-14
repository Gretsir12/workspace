package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
	"workspaces/internal/platform/httpserver" // поправьте путь под ваш проект
	"workspaces/projects/internal/models"
	"workspaces/projects/internal/service"

	"github.com/gorilla/websocket"
)

type WorkspaceHandler struct {
	s       *service.WorkspaceService
	baseCtx context.Context // отменяется при остановке сервера
	stop    context.CancelFunc
	wg      sync.WaitGroup // активные WebSocket-сессии
}

func NewWorkspaceHandler(service *service.WorkspaceService) *WorkspaceHandler {

	ctx, cancel := context.WithCancel(context.Background())

	return &WorkspaceHandler{
		s:       service,
		baseCtx: ctx,
		stop:    cancel,
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	CheckOrigin: func(r *http.Request) bool {
		// TODO: сравнить с разрешённым списком origin из конфига
		return true
	},
}

// Сообщения сервера о ходе сборки: {"type":"status"|"log","data":"..."}.
type progressMsg struct {
	Type string `json:"type"`
	Data string `json:"data"`
}

// Изменение размера терминала (текстовый кадр после ready).
// Если не нужно, удалите вместе с веткой в цикле ниже.
type resizeMsg struct {
	Cols uint `json:"cols"`
	Rows uint `json:"rows"`
}

// wsWriter сериализует записи: gorilla не позволяет писать из нескольких горутин.
type wsWriter struct {
	mu   sync.Mutex
	conn *websocket.Conn
}

func (w *wsWriter) json(v any) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.conn.WriteJSON(v)
}

func (w *wsWriter) binary(p []byte) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.conn.WriteMessage(websocket.BinaryMessage, p)
}

func (w *wsWriter) status(s string) { _ = w.json(progressMsg{Type: "status", Data: s}) }

// fail отправляет ошибку вашим типом httpserver.ErrorResponse.
func (w *wsWriter) fail(message string) {
	_ = w.json(httpserver.ErrorResponse{Message: message, Timestamp: time.Now()})
}

// Write делает wsWriter io.Writer'ом для лога сборки.
func (w *wsWriter) Write(p []byte) (int, error) {
	_ = w.json(progressMsg{Type: "log", Data: string(p)})
	return len(p), nil
}

type frame struct {
	binary bool
	data   []byte
}

func (h *WorkspaceHandler) Open(w http.ResponseWriter, r *http.Request) {
	if h.baseCtx.Err() != nil { // сервер уже останавливается
		httpserver.ResponseError(w, "server is shutting down", http.StatusServiceUnavailable)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	h.wg.Add(1)
	defer h.wg.Done()

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	stop := context.AfterFunc(h.baseCtx, cancel)
	defer stop()

	out := &wsWriter{conn: conn}

	// Единственный читатель сокета. Любая ошибка чтения (в том числе обрыв
	// во время сборки) отменяет ctx и тем самым прерывает сборку.
	in := make(chan frame, 64)
	go func() {
		defer cancel()
		defer close(in)
		for {
			mt, data, err := conn.ReadMessage()
			if err != nil {
				return
			}
			select {
			case in <- frame{binary: mt == websocket.BinaryMessage, data: data}:
			case <-ctx.Done():
				return
			}
		}
	}()

	// 1. Первое сообщение: models.CreateWorkspaceRequest.
	var req models.CreateWorkspaceRequest
	select {
	case f, ok := <-in:
		if !ok {
			return
		}
		if f.binary {
			out.fail("expected json message")
			return
		}
		if err := json.Unmarshal(f.data, &req); err != nil {
			out.fail(fmt.Sprintf("Failed to open workspace: invalid json: %v", err))
			return
		}
	case <-time.After(30 * time.Second):
		out.fail("timeout waiting for workspace request")
		return
	case <-ctx.Done():
		return
	}
	if err := req.Validate(); err != nil {
		out.fail(fmt.Sprintf("Failed to open workspace: %v", err))
		return
	}

	// 2. Сборка: лог летит в сокет через out.
	out.status("building")
	workspace, err := h.s.OpenWorkspace(ctx, &req, out)
	if err != nil {
		out.fail(fmt.Sprintf("Failed to open workspace: %v", err))
		return
	}

	// 3. Контейнер.
	out.status("starting")
	term, err := h.s.StartTerminal(ctx, workspace, req.ProjectId)
	if err != nil {
		out.fail(fmt.Sprintf("Failed to start container: %v", err))
		return
	}
	defer h.s.StopWorkspace(workspace.ID) // выполнится после term.Close()
	defer term.Close()

	out.status("ready")

	// 4. Контейнер -> клиент.
	go func() {
		defer cancel()
		buf := make([]byte, 4096)
		for {
			n, err := term.Read(buf)
			if n > 0 {
				if werr := out.binary(buf[:n]); werr != nil {
					return
				}
			}
			if err != nil {
				out.status("exited")
				return
			}
		}
	}()

	// 5. Клиент -> контейнер: бинарные кадры идут в stdin,
	// текстовые {"cols":..,"rows":..} меняют размер терминала.
	for {
		select {
		case <-ctx.Done():
			return
		case f, ok := <-in:
			if !ok {
				return
			}
			if f.binary {
				if _, err := term.Write(f.data); err != nil {
					return
				}
				continue
			}
			var m resizeMsg
			if json.Unmarshal(f.data, &m) == nil && m.Cols > 0 && m.Rows > 0 {
				_ = term.Resize(ctx, m.Cols, m.Rows)
			}
		}
	}
}

func (h *WorkspaceHandler) Shutdown(ctx context.Context) error {
	h.stop()

	done := make(chan struct{})
	go func() {
		h.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
