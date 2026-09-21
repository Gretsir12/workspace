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
	db := app.GetConnection(cfg)

	projectRepo := repository.NewProjectRepository(cfg, db)
	fileRepo := repository.NewFileRepository(cfg, db)
	projectService := service.NewProjectService(cfg, projectRepo)
	fileService := service.NewFileService(cfg, projectRepo, fileRepo)
	projectHandler := handler.NewProjectHandler(projectService)
	fileHandler := handler.NewFileHandler(fileService)

	app := &app.App{
		Services: app.Services{
			Project: projectService,
			File:    fileService,
		},
		Handlers: app.Handlers{
			Project: projectHandler,
			File:    fileHandler,
		},
	}

	server := api.New(cfg, app.Handlers)
	server.Run()
}
