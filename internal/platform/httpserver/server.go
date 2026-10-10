package httpserver

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/go-chi/httplog/v3"
)

func NewRouter(service string) *chi.Mux {
	r := chi.NewRouter()
	setupLogger(r, service)
	setupCors(r)
	return r
}

func setupLogger(r *chi.Mux, service string) {
	schema := httplog.SchemaECS.Concise(true)

	logger := slog.New(
		slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level:       slog.LevelDebug,
			ReplaceAttr: schema.ReplaceAttr,
		}),
	).With(slog.String("service", service))

	r.Use(httplog.RequestLogger(logger, &httplog.Options{
		Level:         slog.LevelInfo,
		Schema:        schema,
		RecoverPanics: true,
	}))
}

func setupCors(r *chi.Mux) {
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"https://*", "http://*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
	}))
}

// Run блокируется до отмены ctx, затем корректно останавливает сервер.
func Run(ctx context.Context, addr string, h http.Handler) error {
	srv := &http.Server{Addr: addr, Handler: h, ReadHeaderTimeout: 5 * time.Second}

	errCh := make(chan error, 1)
	go func() {
		slog.Info("starting HTTP server", "addr", addr)
		errCh <- srv.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	}
}
