package requtil

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

// Wrap turns a handler into an http.HandlerFunc: an *ApiError is written as {"error": msg} with
// its status, any other error is logged and hidden behind a generic 500, nil is written as 204
// and anything else as 200 JSON.
func Wrap[R any](f func(r *http.Request) (*R, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, err := f(r)
		if err != nil {
			var apiErr *ApiError
			if !errors.As(err, &apiErr) {
				log.Printf("%s %s: %v", r.Method, r.URL.Path, err)
				apiErr = NewError(http.StatusInternalServerError, "internal server error")
			}
			writeJSON(w, apiErr.Status, map[string]string{"error": apiErr.Message})
			return
		}

		if data == nil {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		writeJSON(w, http.StatusOK, data)
	}
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		log.Printf("write json response: %v", err)
	}
}
