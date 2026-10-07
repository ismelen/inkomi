package domain

import "net/http"

type SocketHub interface {
	Register(id string, w http.ResponseWriter, r *http.Request) error
	Send(id string, v any) error
}
