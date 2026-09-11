package models

type CreateProjectRequest struct {
	Title   *string `json:"title"`
	OwnerID *int64  `json:"owner_id"`
}

type Project struct {
	ID    uint32
	User  uint32
	Title string
}
