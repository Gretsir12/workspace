package custom_errors

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"time"
)

type ErrorResponse struct {
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

var (
	ErrProjectNotFound      = errors.New("project not found")
	ErrProjectAlreadyExists = errors.New("project already exists")
	ErrInvalidProject       = errors.New("invalid project")
	ErrInvalidFile          = errors.New("invalid file")
	ErrFileExists           = errors.New("file exists")
	ErrServerError          = errors.New("internal server error")
	ErrInvalidQuery         = errors.New("failed to parse query")
)

func StatusFromError(err error) int {
	switch {
	case errors.Is(err, ErrProjectNotFound):
		return http.StatusBadRequest

	case errors.Is(err, ErrProjectAlreadyExists):
		return http.StatusConflict

	case errors.Is(err, ErrInvalidProject):
		return http.StatusBadRequest

	case errors.Is(err, ErrServerError):
		return http.StatusInternalServerError

	case errors.Is(err, ErrInvalidFile):
		return http.StatusBadRequest

	case errors.Is(err, ErrFileExists):
		return http.StatusBadRequest

	case errors.Is(err, ErrInvalidQuery):
		return http.StatusBadRequest

	case errors.Is(err, io.ErrUnexpectedEOF):
		return http.StatusBadRequest

	default:
		return http.StatusInternalServerError
	}
}

func Error(w http.ResponseWriter, error string, code int) {
	h := w.Header()

	// Delete the Content-Length header, which might be for some other content.
	// Assuming the error string fits in the writer's buffer, we'll figure
	// out the correct Content-Length for it later.
	//
	// We don't delete Content-Encoding, because some middleware sets
	// Content-Encoding: gzip and wraps the ResponseWriter to compress on-the-fly.
	// See https://go.dev/issue/66343.
	h.Del("Content-Length")

	// There might be content type already set, but we reset it to
	// text/plain for the error message.
	h.Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)

	body := ErrorResponse{
		Message:   error,
		Timestamp: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(body); err != nil {
		log.Printf("encode error %s: %v", body, err)
	}
}
