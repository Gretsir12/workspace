package models

import (
	"workspaces/projects/internal/custom_errors"
)

type CreateProjectRequest struct {
	Title        *string      `json:"title"`
	OwnerID      *int64       `json:"owner_id"`
	Dependencies []Dependency `json:"dependencies"`
}

func (r *CreateProjectRequest) Validate() error {
	if r == nil ||
		r.Title == nil || *r.Title == "" ||
		r.OwnerID == nil ||
		r.Dependencies == nil || len(r.Dependencies) == 0 {
		return custom_errors.ErrInvalidProject // или любой другой sentinel error
	}
	return nil
}

type Project struct {
	ID           uint32       `json:"id"`
	User         uint32       `json:"user"`
	Title        string       `json:"title"`
	Path         string       `json:"path"`
	Dependencies []Dependency `json:"dependencies"`

	WorkspaceId uint32 `json:"workspace_id"`
	Files       []File `json:"files"`
}
