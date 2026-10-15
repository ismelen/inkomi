package mocks

import (
	"context"

	"github.com/ismelen/inkomi/back/contracts/jobs"
)

type MockJoinStore struct {
	AddFn func(ctx context.Context, parentId string, total int, msg jobs.Msg) ([]jobs.Msg, error)
}

func (m *MockJoinStore) Add(ctx context.Context, parentId string, total int, msg jobs.Msg) ([]jobs.Msg, error) {
	if m.AddFn != nil {
		return m.AddFn(ctx, parentId, total, msg)
	}
	return nil, nil
}
