package dtos

import "github.com/ismelen/inkomi/back/auth-service/internal/domain/models"

type DropboxProfileResponse struct {
	Linked bool   `json:"linked"`
	Email  string `json:"email,omitempty"`
	Folder string `json:"folder"`
}

// NewDropboxProfileResponse never exposes the tokens of link, which may be nil.
func NewDropboxProfileResponse(link *models.DropboxLink) DropboxProfileResponse {
	if link == nil {
		return DropboxProfileResponse{}
	}
	return DropboxProfileResponse{Linked: true, Email: link.Email, Folder: link.Folder}
}
