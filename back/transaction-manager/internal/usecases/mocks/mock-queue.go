package mocks

import (
	"context"

	"github.com/ismelen/inkomi/back/contracts/jobs"
	"github.com/ismelen/inkomi/back/transaction-manager/internal/domain/ports"
)

type MockQueue struct {
	PublishFn   func(ctx context.Context, subject jobs.Subject, payload []byte) error
	SubscribeFn func(subject jobs.Subject, handler ports.EventHandler) error
}

func (m *MockQueue) Publish(ctx context.Context, subject jobs.Subject, payload []byte) error {
	if m.PublishFn != nil {
		return m.PublishFn(ctx, subject, payload)
	}
	return nil
}

func (m *MockQueue) Subscribe(subject jobs.Subject, handler ports.EventHandler) error {
	if m.SubscribeFn != nil {
		return m.SubscribeFn(subject, handler)
	}
	return nil
}

func (m *MockQueue) Close() error {
	return nil
}
