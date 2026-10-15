package ports

import "net/http"

type ISocketHub interface {
	Register(id int, w http.ResponseWriter, r *http.Request) error
	Send(id int, v any) error
}
