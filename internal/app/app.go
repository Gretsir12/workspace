package app

import (
	"workspaces/projects/internal/handler"
	"workspaces/projects/internal/service"
)

type Handlers struct {
	Project *handler.ProjectHandler
	File    *handler.FileHandler
}

type Services struct {
	Project *service.ProjectService
	File    *service.FileService
}

type App struct {
	Services Services
	Handlers Handlers
}
