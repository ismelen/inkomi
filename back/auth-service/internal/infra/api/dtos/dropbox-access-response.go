package dtos

import "github.com/ismelen/inkomi/back/auth-service/internal/domain/models"

// DropboxAccessResponse is what push-service needs to upload to the Dropbox of a user.
type DropboxAccessResponse struct {
	Token  string `json:"token"`
	Folder string `json:"folder"`
}

func NewDropboxAccessResponse(link *models.DropboxLink) *DropboxAccessResponse {
	return &DropboxAccessResponse{Token: link.Access.Token, Folder: link.Folder}
}
