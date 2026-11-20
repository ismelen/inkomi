package middlewares

import "net/http"

// MaxQueryLength rejects requests whose raw query is longer than n bytes.
func MaxQueryLength(n int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if len(r.URL.RawQuery) > n {
				http.Error(w, "query too long", http.StatusRequestURITooLong)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
