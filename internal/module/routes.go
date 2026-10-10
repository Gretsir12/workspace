package module

import "github.com/go-chi/chi/v5"

func (m *Module) Routes(r chi.Router) {
	r.Get("/proj/{id}", m.Handlers.Project.Get)
	r.Post("/proj", m.Handlers.Project.Create)

	r.Get("/files", m.Handlers.File.Get)
	r.Get("/files/{project_id}", m.Handlers.File.GetAllProjectFiles)
	r.Post("/files", m.Handlers.File.Create)
	r.Put("/files", m.Handlers.File.Update)
}
