package models

type CreateFileRequest struct {
	ProjectID uint32 `json:"project_id"`
	Filname   string `json:"filename"`
}

type File struct {
	Project *Project `json:"project"`
	Filname string   `json:"filename"`
	Path    string   `json:"path"`
}
