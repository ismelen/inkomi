package ports

import "github.com/ismelen/inkomi/back/auth-service/internal/domain/models"

type ITokenIssuer interface {
	IssueAccess(user *models.User) (models.AccessToken, error)
	// NewRefresh returns a random refresh token and the hash that is stored instead of it.
	NewRefresh() (token string, hash string, err error)
	HashRefresh(token string) string
}
