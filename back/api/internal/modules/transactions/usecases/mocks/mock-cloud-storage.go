package mocks

import "time"

type MockCloudStorage struct {
	GetUrlFn         func(id string, filename string, size int64) (string, error)
	GetDownloadUrlFn func(id string, filename string) (string, time.Time, error)
	CheckFn          func(id string, filename string) bool
}

func (m *MockCloudStorage) GetUrl(id string, filename string, size int64) (string, error) {
	if m.GetUrlFn != nil {
		return m.GetUrlFn(id, filename, size)
	}
	return "", nil
}

func (m *MockCloudStorage) GetDownloadUrl(id string, filename string) (string, time.Time, error) {
	if m.GetDownloadUrlFn != nil {
		return m.GetDownloadUrlFn(id, filename)
	}
	return "", time.Time{}, nil
}

func (m *MockCloudStorage) Check(id string, filename string) bool {
	if m.CheckFn != nil {
		return m.CheckFn(id, filename)
	}
	return false
}
