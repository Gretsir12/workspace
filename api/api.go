package api

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"workspaces/projects/internal/app"
	"workspaces/projects/internal/util"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httplog/v3"
)

type API struct {
	router   chi.Router
	handlers app.Handlers
	config   *util.Config
}

func New(config *util.Config, handlers app.Handlers) *API {
	r := chi.NewRouter()

	setupLogger(r)

	a := &API{
		router:   r,
		handlers: handlers,
		config:   config,
	}

	a.setupRoutes()

	return a
}

func setupLogger(r *chi.Mux) {
	schema := httplog.SchemaECS.Concise(true)

	logger := slog.New(
		slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level:       slog.LevelDebug,
			ReplaceAttr: schema.ReplaceAttr,
		}),
	).With(
		slog.String("service", "projects-api"),
	)

	r.Use(httplog.RequestLogger(logger, &httplog.Options{
		Level:         slog.LevelInfo,
		Schema:        schema,
		RecoverPanics: true,
	}))
}

func (a *API) setupRoutes() {
	//projects
	a.router.Get("/api/v1/proj/{id}", a.handlers.Project.Get)
	a.router.Post("/api/v1/proj", a.handlers.Project.Create)

	// files
	a.router.Post("/api/v1/files", a.handlers.File.Create)
	a.router.Get("/api/v1/files", a.handlers.File.Get)
}

func (a *API) Run() error {
	addr := fmt.Sprintf(":%d", a.config.API_PORT)

	slog.Info("starting HTTP server",
		"addr", addr,
	)

	return http.ListenAndServe(addr, a.router)
}
