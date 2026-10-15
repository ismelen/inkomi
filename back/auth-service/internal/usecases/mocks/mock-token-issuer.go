package mocks

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/ismelen/inkomi/back/auth-service/internal/domain/models"
)

// MockTokenIssuer issues "access:<userId>" and "refresh-<n>" tokens hashed as "hash:" + token
// unless its functions are set.
type MockTokenIssuer struct {
	IssueAccessFn func(user *models.User) (models.AccessToken, error)
	NewRefreshFn  func() (string, string, error)
	HashRefreshFn func(token string) string
	refreshCount  atomic.Int32
}

func (m *MockTokenIssuer) IssueAccess(user *models.User) (models.AccessToken, error) {
	if m.IssueAccessFn != nil {
		return m.IssueAccessFn(user)
	}
	return models.AccessToken{Token: fmt.Sprintf("access:%d", user.Id), ExpiresAt: time.Now().Add(time.Minute)}, nil
}

func (m *MockTokenIssuer) NewRefresh() (string, string, error) {
	if m.NewRefreshFn != nil {
		return m.NewRefreshFn()
	}
	token := fmt.Sprintf("refresh-%d", m.refreshCount.Add(1))
	return token, m.HashRefresh(token), nil
}

func (m *MockTokenIssuer) HashRefresh(token string) string {
	if m.HashRefreshFn != nil {
		return m.HashRefreshFn(token)
	}
	return "hash:" + token
}
