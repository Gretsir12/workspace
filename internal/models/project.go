package models

type CreateProjectRequest struct {
	Title   *string `json:"title"`
	OwnerID *int64  `json:"owner_id"`
}

type Project struct {
	ID    uint32 `json:"id"`
	User  uint32 `json:"user"`
	Title string `json:"title"`

	Files []File `json:"files"`
}
