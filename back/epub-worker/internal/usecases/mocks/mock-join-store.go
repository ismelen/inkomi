package mocks

import (
	"context"

	"github.com/ismelen/inkomi/back/common/jobs"
)

type MockJoinStore struct {
	AddFn     func(ctx context.Context, parentId string, total int, msg jobs.Msg) ([]jobs.Msg, error)
	ClearFn   func(ctx context.Context, parentId string) error
	ReleaseFn func(ctx context.Context, parentId string) error
}

func (m *MockJoinStore) Add(ctx context.Context, parentId string, total int, msg jobs.Msg) ([]jobs.Msg, error) {
	if m.AddFn != nil {
		return m.AddFn(ctx, parentId, total, msg)
	}
	return nil, nil
}

func (m *MockJoinStore) Clear(ctx context.Context, parentId string) error {
	if m.ClearFn != nil {
		return m.ClearFn(ctx, parentId)
	}
	return nil
}

func (m *MockJoinStore) Release(ctx context.Context, parentId string) error {
	if m.ReleaseFn != nil {
		return m.ReleaseFn(ctx, parentId)
	}
	return nil
}
