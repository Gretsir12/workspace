package util

import (
	"net/http"
	"strconv"
	"workspaces/projects/internal/custom_errors"
)

func ParseUintFromQuery(name string, r *http.Request) (uint32, error) {
	// parsing file id from query
	fileIDStr := r.URL.Query().Get(name)
	if fileIDStr == "" {
		return 0, custom_errors.ErrInvalidQuery
	}

	fileID, err := strconv.ParseUint(fileIDStr, 10, 32)
	if err != nil {
		return 0, custom_errors.ErrInvalidQuery
	}

	return uint32(fileID), nil
}

func ParseStringFromQuery(name string, r *http.Request) (string, error) {
	// parsing file id from query
	raw := r.URL.Query().Get(name)
	if raw == "" {
		return "", custom_errors.ErrInvalidQuery
	}

	return raw, nil
}
