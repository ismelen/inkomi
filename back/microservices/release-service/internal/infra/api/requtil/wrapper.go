package requtil

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

// Wrap turns a handler into an http.HandlerFunc: an error is written as {"error": msg} with
// the status of an *ApiError (500 otherwise), nil as 204, a *Redirect as 302 and anything else
// as 200 JSON.
func Wrap[R any](f func(r *http.Request) (*R, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, err := f(r)
		if err != nil {
			status := http.StatusInternalServerError
			var apiErr *ApiError
			if errors.As(err, &apiErr) {
				status = apiErr.Status
			} else {
				log.Printf("%s %s: %v", r.Method, r.URL.Path, err)
			}
			writeJSON(w, status, map[string]string{"error": err.Error()})
			return
		}

		if data == nil {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if redirect, ok := any(data).(*Redirect); ok {
			http.Redirect(w, r, redirect.URL, http.StatusFound)
			return
		}
		writeJSON(w, http.StatusOK, data)
	}
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		log.Printf("write json response: %v", err)
	}
}
