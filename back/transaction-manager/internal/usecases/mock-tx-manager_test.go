package usecases_test

import (
	"context"
)

type MockTxManager struct {
	ExecuteTxFn func(ctx context.Context, fn func(ctx context.Context) error) error
}

func (m *MockTxManager) ExecuteTx(ctx context.Context, fn func(ctx context.Context) error) error {
	if m.ExecuteTxFn != nil {
		return m.ExecuteTxFn(ctx, fn)
	}
	return fn(ctx)
}

