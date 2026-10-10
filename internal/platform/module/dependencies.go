package module

import (
	"workspaces/internal/platform/config"

	"github.com/go-pg/pg/v10"
)

type Deps struct {
	Cfg *config.Config
	DB  *pg.DB
}
