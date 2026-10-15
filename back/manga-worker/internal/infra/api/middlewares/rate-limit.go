package middlewares

import (
	"context"
	"math"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// IPRateLimiter limits requests per client IP with a token bucket per IP.
type IPRateLimiter struct {
	limit      rate.Limit
	burst      int
	retryAfter string
	trustProxy bool

	mu       sync.Mutex
	visitors map[string]*visitor
}

// NewIPRateLimiter allows perMinute requests per IP with bursts of burst. With trustProxy the IP
// is read from X-Forwarded-For, which is only safe behind a proxy that overwrites it.
func NewIPRateLimiter(perMinute, burst int, trustProxy bool) *IPRateLimiter {
	return &IPRateLimiter{
		limit:      rate.Limit(float64(perMinute) / 60),
		burst:      burst,
		retryAfter: strconv.Itoa(int(math.Ceil(60 / float64(perMinute)))),
		trustProxy: trustProxy,
		visitors:   make(map[string]*visitor),
	}
}

func (l *IPRateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !l.allow(ClientIP(r, l.trustProxy)) {
			w.Header().Set("Retry-After", l.retryAfter)
			http.Error(w, "too many requests, try again later", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// Start forgets the IPs not seen for interval every interval, until ctx is done.
func (l *IPRateLimiter) Start(ctx context.Context, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case now := <-ticker.C:
				l.prune(now.Add(-interval))
			}
		}
	}()
}

func (l *IPRateLimiter) allow(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	v, ok := l.visitors[ip]
	if !ok {
		v = &visitor{limiter: rate.NewLimiter(l.limit, l.burst)}
		l.visitors[ip] = v
	}
	v.lastSeen = time.Now()
	return v.limiter.Allow()
}

func (l *IPRateLimiter) prune(before time.Time) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for ip, v := range l.visitors {
		if v.lastSeen.Before(before) {
			delete(l.visitors, ip)
		}
	}
}

// ClientIP returns the first X-Forwarded-For hop when trustProxy is set, or the remote address.
func ClientIP(r *http.Request, trustProxy bool) string {
	if trustProxy {
		if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
			first, _, _ := strings.Cut(forwarded, ",")
			if ip := strings.TrimSpace(first); ip != "" {
				return ip
			}
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
