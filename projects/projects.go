package projects

import (
	"context"
	"fmt"

	"github.com/go-chi/chi/v5"
	"github.com/go-pg/pg/v10"

	"workspaces/internal/platform/config"
	"workspaces/projects/internal/app"
)

type Module struct{ app *app.App }

func New(ctx context.Context, cfg *config.Config, db *pg.DB) (*Module, error) {
	a, err := app.New(ctx, cfg, db)
	if err != nil {
		return nil, fmt.Errorf("register project app: %w", err)
	}
	return &Module{app: a}, nil
}

func (m *Module) Routes(r chi.Router) { m.app.Routes(r) }
