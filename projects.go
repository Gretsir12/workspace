package projects

import (
	platform "workspaces/internal/platform/module"
	projectsmodule "workspaces/projects/internal/module"
)

func New(deps platform.Deps) platform.Module {
	return projectsmodule.New(deps)
}
