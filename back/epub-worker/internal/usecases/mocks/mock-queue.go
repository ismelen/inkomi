package mocks

import (
	"context"

	"github.com/ismelen/inkomi/back/epub-worker/internal/domain/ports"
)

type MockQueue struct {
	PublishFn   func(ctx context.Context, subject string, payload any) error
	SubscribeFn func(subject string, handler ports.EventHandler) error
}

func (m *MockQueue) Publish(ctx context.Context, subject string, payload any) error {
	if m.PublishFn != nil {
		return m.PublishFn(ctx, subject, payload)
	}
	return nil
}

func (m *MockQueue) Subscribe(subject string, handler ports.EventHandler) error {
	if m.SubscribeFn != nil {
		return m.SubscribeFn(subject, handler)
	}
	return nil
}

func (m *MockQueue) Close() error {
	return nil
}
