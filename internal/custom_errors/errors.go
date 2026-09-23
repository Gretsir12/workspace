package custom_errors

import (
	"errors"
	"net/http"
)

var (
	ErrProjectNotFound      = errors.New("project not found")
	ErrProjectAlreadyExists = errors.New("project already exists")
	ErrInvalidProject       = errors.New("invalid project")
	ErrServerError          = errors.New("internal server error")
)

func StatusFromError(err error) int {
	switch {
	case errors.Is(err, ErrProjectNotFound):
		return http.StatusNotFound

	case errors.Is(err, ErrProjectAlreadyExists):
		return http.StatusConflict

	case errors.Is(err, ErrInvalidProject):
		return http.StatusBadRequest

	case errors.Is(err, ErrServerError):
		return http.StatusInternalServerError

	default:
		return http.StatusInternalServerError
	}
}
