package dtos

import "github.com/ismelen/inkomi/back/auth-service/internal/domain/models"

type ProfileResponse struct {
	Id            int                    `json:"id"`
	Email         string                 `json:"email,omitempty"`
	EmailVerified bool                   `json:"emailVerified"`
	HasPassword   bool                   `json:"hasPassword"`
	Providers     []models.Provider      `json:"providers"`
	Dropbox       DropboxProfileResponse `json:"dropbox"`
}

func NewProfileResponse(profile *models.Profile) *ProfileResponse {
	providers := profile.Providers
	if providers == nil {
		providers = []models.Provider{}
	}
	return &ProfileResponse{
		Id:            profile.User.Id,
		Email:         profile.User.Email,
		EmailVerified: profile.User.EmailVerified,
		HasPassword:   profile.User.HasPassword(),
		Providers:     providers,
		Dropbox:       NewDropboxProfileResponse(profile.Dropbox),
	}
}
