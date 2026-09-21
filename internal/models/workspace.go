package models

// dependency name from https://search.nixos.org/packages
type Dependency string

type Workspace struct {
	ID           uint32
	Dependencies []Dependency
}
