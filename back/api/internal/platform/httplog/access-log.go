package httplog

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

// AccessLog logs one line per request: method, path with its query, status, size, duration and
// client ip. skip lists paths left out, such as the healthcheck the container polls.
func AccessLog(skip ...string) func(http.Handler) http.Handler {
	skipped := make(map[string]bool, len(skip))
	for _, path := range skip {
		skipped[path] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if skipped[r.URL.Path] {
				next.ServeHTTP(w, r)
				return
			}

			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			start := time.Now()
			next.ServeHTTP(ww, r)

			status := ww.Status()
			if status == 0 {
				// a hijacked connection (WebSocket) or a handler that wrote nothing
				status = http.StatusSwitchingProtocols
				if r.Header.Get("Upgrade") == "" {
					status = http.StatusOK
				}
			}
			log.Printf("http %s %s %d %dB %s %s",
				r.Method, r.URL.RequestURI(), status, ww.BytesWritten(),
				time.Since(start).Round(time.Millisecond), clientIP(r))
		})
	}
}

// clientIP prefers the address Caddy forwards.
func clientIP(r *http.Request) string {
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return ip
	}
	return r.RemoteAddr
}
