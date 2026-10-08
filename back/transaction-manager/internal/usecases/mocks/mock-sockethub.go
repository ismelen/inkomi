package mocks

import (
	"net/http"
)

type MockSocketHub struct {
	RegisterFn func(id int, w http.ResponseWriter, r *http.Request) error
	SendFn     func(id int, v any) error
}

func (m *MockSocketHub) Register(id int, w http.ResponseWriter, r *http.Request) error {
	if m.RegisterFn != nil {
		return m.RegisterFn(id, w, r)
	}
	return nil
}

func (m *MockSocketHub) Send(id int, v any) error {
	if m.SendFn != nil {
		return m.SendFn(id, v)
	}
	return nil
}

