package models

import(
	"workspaces/projects/internal/custom_errors"
)

type CreateFileRequest struct {
	ProjectID uint32 `json:"project_id"`
	Filename  string `json:"filename"`
}

func (self *CreateFileRequest) Validate() error {
	if self == nil ||
		self.Filename == "" {
		return custom_errors.ErrInvalidFile
	}	
	return nil
}

type GetFileRequest struct {
	ProjectID uint32 `json:"project_id"`
	Path  string `json:"path"`
}

type EditFileRequest struct {
	ProjectID uint32 `json:"project_id"`
	Path  string `json:"path"`
	Text string `json:text`
}

type File struct {
	ProjectID uint32 `json:"project"`
	Path      string `json:"path"`
	Text      string `json:text`
}
