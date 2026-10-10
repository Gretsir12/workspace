package module

import (
	"context"
	"workspaces/internal/platform/config"
	"workspaces/projects/internal/handler"

	"github.com/go-pg/pg/v10"
)

type Handlers struct {
	Project *handler.ProjectHandler
	File    *handler.FileHandler
}

type Module struct {
	Ctx context.Context
	Cfg *config.Config
	Db  *pg.DB

	Handlers *Handlers
}
