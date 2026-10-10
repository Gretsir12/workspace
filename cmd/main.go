package main

import (
	"context"
	"fmt"
	"log"
	"os/signal"
	"syscall"

	"workspaces/internal/platform/config"
	"workspaces/internal/platform/database"
	"workspaces/internal/platform/httpserver"
	"workspaces/projects"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Load config errror: ", err)
	}

	db := database.GetConnection(&cfg)
	defer db.Close()

	projectsMod, err := projects.New(ctx, &cfg, db)
	if err != nil {
		log.Fatal(err)
	}

	r := httpserver.NewRouter("projects-api")
	r.Route("/api/v1", projectsMod.Routes)

	if err := httpserver.Run(ctx, fmt.Sprintf(":%d", cfg.APIPort), r); err != nil {
		log.Fatal(err)
	}
}
