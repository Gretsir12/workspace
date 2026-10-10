package app

import "github.com/go-chi/chi/v5"

func (a *App) Routes(r chi.Router) {
	r.Get("/proj/{id}", a.Handlers.Project.Get)
	r.Post("/proj", a.Handlers.Project.Create)

	r.Get("/files", a.Handlers.File.Get)
	r.Get("/files/{project_id}", a.Handlers.File.GetAllProjectFiles)
	r.Post("/files", a.Handlers.File.Create)
	r.Put("/files", a.Handlers.File.Update)

}
