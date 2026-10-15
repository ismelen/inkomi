package mocks

import "context"

// MockTxManager runs fn directly unless ExecuteTxFn is set.
type MockTxManager struct {
	ExecuteTxFn func(ctx context.Context, fn func(ctx context.Context) error) error
}

func (m *MockTxManager) ExecuteTx(ctx context.Context, fn func(ctx context.Context) error) error {
	if m.ExecuteTxFn != nil {
		return m.ExecuteTxFn(ctx, fn)
	}
	return fn(ctx)
}
