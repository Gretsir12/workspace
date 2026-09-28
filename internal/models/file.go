package models

type CreateFileRequest struct {
	ProjectID uint32 `json:"project_id"`
	Filename  string `json:"filename"`
}

type GetFileRequest struct {
	ProjectID uint32 `json:"project_id"`
	Filename  string `json:"filename"`
}

type File struct {
	ProjectID uint32 `json:"project"`
	Path      string `json:"path"`
}
