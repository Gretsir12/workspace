package models

import (
	"errors"
	"fmt"
	"regexp"
	"time"
)

const (
	maxDependencies  = 50
	maxDependencyLen = 100
)

var ErrInvalidRequest = errors.New("invalid request")

// nixpkgs attribute name: dot-separated segments (e.g. python311Packages.requests)
var depRe = regexp.MustCompile(`^[a-zA-Z0-9_][a-zA-Z0-9_+-]*(\.[a-zA-Z0-9_][a-zA-Z0-9_+-]*)*$`)

type CreateWorkspaceRequest struct {
	UserId       uint32   `json:"user_id"`
	ProjectId    uint32   `json:"project_id"`
	Dependencies []string `json:"dependencies"` // dependency name from https://search.nixos.org/packages
}

func validateDependency(s string) error {
	switch {
	case s == "":
		return errors.New("empty name")
	case len(s) > maxDependencyLen:
		return fmt.Errorf("%q...: too long (max %d)", s[:20], maxDependencyLen)
	case !depRe.MatchString(s):
		return fmt.Errorf("%q: invalid characters", s)
	}
	return nil
}

func (r *CreateWorkspaceRequest) Validate() error {
	var errs []error

	if r.UserId == 0 {
		errs = append(errs, errors.New("user_id is required"))
	}
	if r.ProjectId == 0 {
		errs = append(errs, errors.New("project_id is required"))
	}

	switch {
	case len(r.Dependencies) == 0:
		errs = append(errs, errors.New("dependencies must not be empty"))
	case len(r.Dependencies) > maxDependencies:
		errs = append(errs, fmt.Errorf("too many dependencies (max %d)", maxDependencies))
	default:
		seen := make(map[string]struct{}, len(r.Dependencies))
		for i, d := range r.Dependencies {
			if err := validateDependency(d); err != nil {
				errs = append(errs, fmt.Errorf("dependencies[%d]: %w", i, err))
				continue
			}
			if _, dup := seen[d]; dup {
				errs = append(errs, fmt.Errorf("dependencies[%d]: duplicate %q", i, d))
			}
			seen[d] = struct{}{}
		}
	}

	if len(errs) == 0 {
		return nil
	}
	return fmt.Errorf("%w: %w", ErrInvalidRequest, errors.Join(errs...))
}

type Image struct {
	tableName struct{} `pg:"workspace_images"`

	Hash       string    `pg:"hash,pk,type:char(64)"`
	NixpkgsRev string    `pg:"nixpkgs_rev,notnull"`
	Deps       []string  `pg:"deps,array,notnull"`
	Image      string    `pg:"image,notnull"`  // "registry.local/ws:<hash>"
	Status     string    `pg:"status,notnull"` // building | ready | failed
	CreatedAt  time.Time `pg:"created_at,notnull,default:now()"`
	UpdatedAt  time.Time `pg:"updated_at,notnull,default:now()"`
	LastUsedAt time.Time `pg:"last_used_at,notnull,default:now()"`
}

type Workspace struct {
	tableName struct{} `pg:"workspaces"`

	ID          uint32
	UserID      uint32    `pg:"user_id,notnull"`
	ImageHash   string    `pg:"image_hash,notnull,type:char(64)"`
	Image       *Image    `pg:"rel:has-one"`
	ContainerID string    `pg:"container_id"`   // пусто, пока контейнер не создан
	Status      string    `pg:"status,notnull"` // pending | running | stopped | failed
	ExpiresAt   time.Time `pg:"expires_at,notnull"`
	CreatedAt   time.Time `pg:"created_at,notnull,default:now()"`
}
