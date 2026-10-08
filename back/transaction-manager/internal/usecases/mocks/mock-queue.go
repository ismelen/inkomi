package mocks

import (
	"context"

	"github.com/ismelen/inkomi/back/transaction-manager/internal/domain/ports"
)

type MockQueue struct {
	PublishFn func(ctx context.Context, subject string, payload []byte) error
}

func (m *MockQueue) Publish(ctx context.Context, subject string, payload []byte) error {
	if m.PublishFn != nil {
		return m.PublishFn(ctx, subject, payload)
	}
	return nil
}

func (m *MockQueue) Subscribe(ctx context.Context, subject string, handler ports.EventHandler) error {
	return nil
}

func (m *MockQueue) Close() error {
	return nil
}

