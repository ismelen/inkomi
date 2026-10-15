package mocks

// MockPasswordHasher hashes as "hash:" + password unless its functions are set.
type MockPasswordHasher struct {
	HashFn   func(password string) (string, error)
	VerifyFn func(password, hash string) (bool, error)
}

func (m *MockPasswordHasher) Hash(password string) (string, error) {
	if m.HashFn != nil {
		return m.HashFn(password)
	}
	return "hash:" + password, nil
}

func (m *MockPasswordHasher) Verify(password, hash string) (bool, error) {
	if m.VerifyFn != nil {
		return m.VerifyFn(password, hash)
	}
	return hash == "hash:"+password, nil
}
