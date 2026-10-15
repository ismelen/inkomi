package dtos

type DropboxLoginRequest struct {
	RefreshToken string `json:"refreshToken"`
	DeviceId     string `json:"deviceId"`
}
