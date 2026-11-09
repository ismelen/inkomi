package httplog_test

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/ismelen/inkomi/back/monolith/internal/platform/httplog"
	"github.com/stretchr/testify/assert"
)

// The tests redirect the global logger, so they don't run in parallel.

func TestAccessLog_Request_LogsMethodPathStatusAndSize(t *testing.T) {
	// Arrange
	var buf bytes.Buffer
	log.SetOutput(&buf)
	t.Cleanup(func() { log.SetOutput(os.Stderr) })
	handler := httplog.AccessLog()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("hello"))
	}))
	req := httptest.NewRequest(http.MethodGet, "/preview/processed?gamma=1.2", nil)
	req.Header.Set("X-Forwarded-For", "1.2.3.4")

	// Act
	handler.ServeHTTP(httptest.NewRecorder(), req)

	// Assert
	assert.Contains(t, buf.String(), "http GET /preview/processed?gamma=1.2 201 5B")
	assert.Contains(t, buf.String(), "1.2.3.4")
}

func TestAccessLog_SkippedPath_LogsNothing(t *testing.T) {
	// Arrange
	var buf bytes.Buffer
	log.SetOutput(&buf)
	t.Cleanup(func() { log.SetOutput(os.Stderr) })
	handler := httplog.AccessLog("/healthz")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	// Act
	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/healthz", nil))

	// Assert
	assert.Empty(t, buf.String())
}
