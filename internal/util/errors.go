package util

import (
	"errors"
	"net/http"
)

var (
	ErrProjectNotFound      = errors.New("project not found")
	ErrProjectAlreadyExists = errors.New("project already exists")
	ErrInvalidProject       = errors.New("invalid project")
)

func StatusFromError(err error) int {
	switch {
	case errors.Is(err, ErrProjectNotFound):
		return http.StatusNotFound

	case errors.Is(err, ErrProjectAlreadyExists):
		return http.StatusConflict

	case errors.Is(err, ErrInvalidProject):
		return http.StatusBadRequest

	default:
		return http.StatusInternalServerError
	}
}
