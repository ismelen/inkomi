package mocks

import (
	"context"

	"github.com/ismelen/inkomi/back/push-service/internal/domain/models"
)

type MockPushNotifier struct {
	SendFn func(ctx context.Context, token string, message models.PushMessage) error
}

func (m *MockPushNotifier) Send(ctx context.Context, token string, message models.PushMessage) error {
	if m.SendFn != nil {
		return m.SendFn(ctx, token, message)
	}
	return nil
}
