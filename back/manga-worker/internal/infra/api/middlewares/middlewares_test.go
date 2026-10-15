package middlewares_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ismelen/inkomi/back/manga-worker/internal/infra/api/middlewares"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var ok = http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })

func get(handler http.Handler, remoteAddr string, headers map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/preview/processed", nil)
	req.RemoteAddr = remoteAddr
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

func TestIPRateLimiter_Middleware_OverBurst_Returns429(t *testing.T) {
	t.Parallel()

	// Arrange
	handler := middlewares.NewIPRateLimiter(6, 3, false).Middleware(ok)

	// Act
	var codes []int
	for range 4 {
		codes = append(codes, get(handler, "1.1.1.1:1000", nil).Code)
	}
	other := get(handler, "2.2.2.2:1000", nil)
	limited := get(handler, "1.1.1.1:2000", nil)

	// Assert
	assert.Equal(t, []int{200, 200, 200, 429}, codes)
	assert.Equal(t, http.StatusOK, other.Code)
	assert.Equal(t, http.StatusTooManyRequests, limited.Code)
	assert.Equal(t, "10", limited.Header().Get("Retry-After"))
}

func TestIPRateLimiter_Middleware_ForwardedFor_OnlyTrustedBehindProxy(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		trustProxy bool
		want       []int
	}{
		{"trusted: each forwarded ip has its own bucket", true, []int{200, 200}},
		{"untrusted: the header is ignored", false, []int{200, 429}},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			handler := middlewares.NewIPRateLimiter(6, 1, c.trustProxy).Middleware(ok)

			// Act
			first := get(handler, "10.0.0.1:1", map[string]string{"X-Forwarded-For": "1.1.1.1, 10.0.0.1"})
			second := get(handler, "10.0.0.1:1", map[string]string{"X-Forwarded-For": "2.2.2.2, 10.0.0.1"})

			// Assert
			assert.Equal(t, c.want, []int{first.Code, second.Code})
		})
	}
}

func TestIPRateLimiter_Start_ForgetsIdleIPs(t *testing.T) {
	t.Parallel()

	// Arrange
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	limiter := middlewares.NewIPRateLimiter(1, 1, false)
	handler := limiter.Middleware(ok)
	require.Equal(t, http.StatusOK, get(handler, "1.1.1.1:1", nil).Code)
	require.Equal(t, http.StatusTooManyRequests, get(handler, "1.1.1.1:1", nil).Code)

	// Act
	limiter.Start(ctx, 10*time.Millisecond)

	// Assert
	assert.Eventually(t, func() bool {
		return get(handler, "1.1.1.1:1", nil).Code == http.StatusOK
	}, 2*time.Second, 50*time.Millisecond)
}

func TestConcurrencyLimit_Full_Returns503(t *testing.T) {
	t.Parallel()

	// Arrange
	release := make(chan struct{})
	started := make(chan struct{})
	blocking := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		started <- struct{}{}
		<-release
		w.WriteHeader(http.StatusOK)
	})
	handler := middlewares.ConcurrencyLimit(1, 20*time.Millisecond)(blocking)

	var wg sync.WaitGroup
	var first *httptest.ResponseRecorder
	wg.Go(func() { first = get(handler, "1.1.1.1:1", nil) })
	<-started

	// Act
	second := get(handler, "1.1.1.1:1", nil)
	close(release)
	wg.Wait()

	// Assert
	assert.Equal(t, http.StatusOK, first.Code)
	assert.Equal(t, http.StatusServiceUnavailable, second.Code)
	assert.NotEmpty(t, second.Header().Get("Retry-After"))
}

func TestMaxQueryLength_LongQuery_Returns414(t *testing.T) {
	t.Parallel()

	// Arrange
	handler := middlewares.MaxQueryLength(10)(ok)
	short := httptest.NewRequest(http.MethodGet, "/preview?a=1", nil)
	long := httptest.NewRequest(http.MethodGet, "/preview?a="+strings.Repeat("x", 20), nil)

	// Act
	shortRec, longRec := httptest.NewRecorder(), httptest.NewRecorder()
	handler.ServeHTTP(shortRec, short)
	handler.ServeHTTP(longRec, long)

	// Assert
	assert.Equal(t, http.StatusOK, shortRec.Code)
	assert.Equal(t, http.StatusRequestURITooLong, longRec.Code)
}

func TestTimeout_SetsDeadline(t *testing.T) {
	t.Parallel()

	// Arrange
	var deadline time.Time
	var hasDeadline bool
	handler := middlewares.Timeout(time.Minute)(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		deadline, hasDeadline = r.Context().Deadline()
	}))

	// Act
	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

	// Assert
	require.True(t, hasDeadline)
	assert.WithinDuration(t, time.Now().Add(time.Minute), deadline, 5*time.Second)
}
