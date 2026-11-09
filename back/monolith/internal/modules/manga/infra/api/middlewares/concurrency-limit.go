package middlewares

import (
	"net/http"
	"time"
)

// ConcurrencyLimit lets at most n requests run at once. A request waits up to wait for a free
// slot and is answered with 503 otherwise.
func ConcurrencyLimit(n int, wait time.Duration) func(http.Handler) http.Handler {
	slots := make(chan struct{}, n)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			timer := time.NewTimer(wait)
			defer timer.Stop()

			select {
			case slots <- struct{}{}:
			case <-timer.C:
				w.Header().Set("Retry-After", "5")
				http.Error(w, "server busy, try again later", http.StatusServiceUnavailable)
				return
			case <-r.Context().Done():
				return
			}
			defer func() { <-slots }()

			next.ServeHTTP(w, r)
		})
	}
}
