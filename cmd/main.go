package main

import (
	"workspaces/projects/api"
	"workspaces/projects/internal/app"
	"workspaces/projects/internal/handler"
	"workspaces/projects/internal/repository"
	"workspaces/projects/internal/service"
	"workspaces/projects/internal/util"
)

func main() {
	cfg := util.NewConfig()
	// db := database.New(cfg)

	projectRepo := repository.NewProjectRepository(cfg)
	projectService := service.NewProjectService(cfg, projectRepo)
	projectHandler := handler.NewProjectHandler(projectService)

	app := &app.App{
		Services: app.Services{
			Project: projectService,
		},
		Handlers: app.Handlers{
			Project: projectHandler,
		},
	}

	server := api.New(cfg, app.Handlers)
	server.Run()
}
