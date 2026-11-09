package dtos

type GoogleLoginRequest struct {
	IdToken  string `json:"idToken"`
	DeviceId string `json:"deviceId"`
}
