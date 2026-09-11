package app

import (
	"workspaces/projects/internal/handler"
	"workspaces/projects/internal/service"
)

type Handlers struct {
	Project *handler.ProjectHandler
}

type Services struct {
	Project *service.ProjectService
}

type App struct {
	Services Services
	Handlers Handlers
}
