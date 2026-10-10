package models

// dependency name from https://search.nixos.org/packages
type Dependency string

type Workspace struct {
	ID           uint32
	Dependencies []Dependency
}

type CreateWorkspaceRequest struct {
	UserId       uint32       `json:"user_id"`
	ProjectId    uint32       `json:"project_id"`
	Dependencies []Dependency `json:"dependecies"`
}
