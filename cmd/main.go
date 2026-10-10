package main

import (
	"context"
	"fmt"
	"log"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"

	"workspaces/internal/platform/config"
	"workspaces/internal/platform/database"
	"workspaces/internal/platform/httpserver"
	"workspaces/internal/platform/module"
	"workspaces/projects"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Load config error: ", err)
	}

	db := database.GetConnection(&cfg)
	defer db.Close()

	deps := module.Deps{Cfg: &cfg, DB: db}

	modules := []module.Module{
		projects.New(deps),
	}

	r := httpserver.NewRouter("projects-api")
	r.Route("/api/v1", func(r chi.Router) {
		for _, m := range modules {
			log.Printf("registering module: %s", m.Name())
			m.Routes(r)
		}
	})

	if err := httpserver.Run(ctx, fmt.Sprintf(":%d", cfg.APIPort), r); err != nil {
		log.Fatal(err)
	}
}
