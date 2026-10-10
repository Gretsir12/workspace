package module

import "github.com/go-chi/chi/v5"

type Module interface {
	Name() string
	Routes(r chi.Router)
}
