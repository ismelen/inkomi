package mocks

import (
	"context"

	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/domain/models"
)

type MockSignalRepository struct {
	AddSignalFn            func(ctx context.Context, signal *models.Signal) (bool, error)
	RemoveWishlistExceptFn func(ctx context.Context, refs map[string]bool) ([]int, error)
	ListPendingFn          func(ctx context.Context, limit, maxAttempts int) ([]models.Signal, error)
	ResolveFn              func(ctx context.Context, id int64, status models.SignalStatus, matchedKey string) error
	ListMatchedFn          func(ctx context.Context, userId int) ([]models.Signal, error)
	ListUsersFn            func(ctx context.Context) ([]int, error)
	MarkStaleFn            func(ctx context.Context, userIds ...int) error
	TakeStaleFn            func(ctx context.Context) ([]int, error)
	DismissFn              func(ctx context.Context, userId int, workKey string) error
	ListDismissedFn        func(ctx context.Context, userId int) (map[string]bool, error)
	GetStateFn             func(ctx context.Context, name string) (string, error)
	SetStateFn             func(ctx context.Context, name, value string) error
}

func (m *MockSignalRepository) AddSignal(ctx context.Context, signal *models.Signal) (bool, error) {
	if m.AddSignalFn != nil {
		return m.AddSignalFn(ctx, signal)
	}
	return true, nil
}

func (m *MockSignalRepository) RemoveWishlistExcept(ctx context.Context, refs map[string]bool) ([]int, error) {
	if m.RemoveWishlistExceptFn != nil {
		return m.RemoveWishlistExceptFn(ctx, refs)
	}
	return nil, nil
}

func (m *MockSignalRepository) ListPending(ctx context.Context, limit, maxAttempts int) ([]models.Signal, error) {
	if m.ListPendingFn != nil {
		return m.ListPendingFn(ctx, limit, maxAttempts)
	}
	return nil, nil
}

func (m *MockSignalRepository) Resolve(ctx context.Context, id int64, status models.SignalStatus, matchedKey string) error {
	if m.ResolveFn != nil {
		return m.ResolveFn(ctx, id, status, matchedKey)
	}
	return nil
}

func (m *MockSignalRepository) ListMatched(ctx context.Context, userId int) ([]models.Signal, error) {
	if m.ListMatchedFn != nil {
		return m.ListMatchedFn(ctx, userId)
	}
	return nil, nil
}

func (m *MockSignalRepository) ListUsers(ctx context.Context) ([]int, error) {
	if m.ListUsersFn != nil {
		return m.ListUsersFn(ctx)
	}
	return nil, nil
}

func (m *MockSignalRepository) MarkStale(ctx context.Context, userIds ...int) error {
	if m.MarkStaleFn != nil {
		return m.MarkStaleFn(ctx, userIds...)
	}
	return nil
}

func (m *MockSignalRepository) TakeStale(ctx context.Context) ([]int, error) {
	if m.TakeStaleFn != nil {
		return m.TakeStaleFn(ctx)
	}
	return nil, nil
}

func (m *MockSignalRepository) Dismiss(ctx context.Context, userId int, workKey string) error {
	if m.DismissFn != nil {
		return m.DismissFn(ctx, userId, workKey)
	}
	return nil
}

func (m *MockSignalRepository) ListDismissed(ctx context.Context, userId int) (map[string]bool, error) {
	if m.ListDismissedFn != nil {
		return m.ListDismissedFn(ctx, userId)
	}
	return map[string]bool{}, nil
}

func (m *MockSignalRepository) GetState(ctx context.Context, name string) (string, error) {
	if m.GetStateFn != nil {
		return m.GetStateFn(ctx, name)
	}
	return "", nil
}

func (m *MockSignalRepository) SetState(ctx context.Context, name, value string) error {
	if m.SetStateFn != nil {
		return m.SetStateFn(ctx, name, value)
	}
	return nil
}
