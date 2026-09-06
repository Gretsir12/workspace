package main

import (
	"workspaces/projects/api"
	"workspaces/projects/internal/handler"
	"workspaces/projects/internal/repository"
	"workspaces/projects/internal/service"
)

func main() {

	config := api.NewConfig()
	projectRepository := repository.NewProjectRepository(config)
	projectService := service.NewProjectService(projectRepository)
	projectHandler := handler.NewProjectHandler(projectService)

}
