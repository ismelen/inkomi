package dtos

import (
	"time"

	"github.com/ismelen/inkomi/back/auth-service/internal/domain/models"
)

type TokensResponse struct {
	AccessToken           string    `json:"accessToken"`
	AccessTokenExpiresAt  time.Time `json:"accessTokenExpiresAt"`
	RefreshToken          string    `json:"refreshToken"`
	RefreshTokenExpiresAt time.Time `json:"refreshTokenExpiresAt"`
}

func NewTokensResponse(pair *models.TokenPair) *TokensResponse {
	return &TokensResponse{
		AccessToken:           pair.AccessToken,
		AccessTokenExpiresAt:  pair.AccessExpiresAt.UTC(),
		RefreshToken:          pair.RefreshToken,
		RefreshTokenExpiresAt: pair.RefreshExpiresAt.UTC(),
	}
}
