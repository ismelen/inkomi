package mocks

import (
	"context"

	"github.com/ismelen/inkomi/back/common/jobrun"
)

type MockQueue struct {
	PublishFn   func(ctx context.Context, subject string, payload any) error
	SubscribeFn func(subject string, handler jobrun.DeliveryHandler) error
}

func (m *MockQueue) Publish(ctx context.Context, subject string, payload any) error {
	if m.PublishFn != nil {
		return m.PublishFn(ctx, subject, payload)
	}
	return nil
}

func (m *MockQueue) Subscribe(subject string, handler jobrun.DeliveryHandler) error {
	if m.SubscribeFn != nil {
		return m.SubscribeFn(subject, handler)
	}
	return nil
}

func (m *MockQueue) Close() error {
	return nil
}
