package mocks

import "context"

type MockSynopsisLookup struct {
	SynopsisFn func(ctx context.Context, workKey, lang string) (string, error)
}

func (m *MockSynopsisLookup) Synopsis(ctx context.Context, workKey, lang string) (string, error) {
	if m.SynopsisFn != nil {
		return m.SynopsisFn(ctx, workKey, lang)
	}
	return "", nil
}
