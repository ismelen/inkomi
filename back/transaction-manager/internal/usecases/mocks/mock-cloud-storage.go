package mocks

type MockCloudStorage struct {
	GetUrlFn func(id string, ext string) (string, error)
	CheckFn  func(id string, ext string) bool
}

func (m *MockCloudStorage) GetUrl(id string, ext string) (string, error) {
	if m.GetUrlFn != nil {
		return m.GetUrlFn(id, ext)
	}
	return "", nil
}

func (m *MockCloudStorage) Check(id string, ext string) bool {
	if m.CheckFn != nil {
		return m.CheckFn(id, ext)
	}
	return false
}

