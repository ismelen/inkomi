package embedding_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/infra/embedding"
)

func TestModelFiles_Ensure_MissingFiles_DownloadsThemOnce(t *testing.T) {
	t.Parallel()

	// Arrange
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		_, _ = w.Write([]byte("content of " + r.URL.Path))
	}))
	t.Cleanup(server.Close)
	dir := t.TempDir()
	model := embedding.NewModelFiles(dir, server.URL+"/", "onnx/model_quantized.onnx")

	// Act
	got, err := model.Ensure(context.Background())
	require.NoError(t, err)
	_, err = model.Ensure(context.Background())
	require.NoError(t, err)

	// Assert
	assert.Equal(t, dir, got)
	assert.Equal(t, int32(5), requests.Load(), "four tokenizer files and the model, once")
	content, err := os.ReadFile(filepath.Join(dir, "model_quantized.onnx"))
	require.NoError(t, err)
	assert.Equal(t, "content of /onnx/model_quantized.onnx", string(content))
	assert.Equal(t, "model_quantized.onnx", model.OnnxFilename())
}

func TestModelFiles_Ensure_ServerError_LeavesNoPartialFile(t *testing.T) {
	t.Parallel()

	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	t.Cleanup(server.Close)
	dir := t.TempDir()

	// Act
	_, err := embedding.NewModelFiles(dir, server.URL, "model.onnx").Ensure(context.Background())

	// Assert
	require.Error(t, err)
	entries, readErr := os.ReadDir(dir)
	require.NoError(t, readErr)
	assert.Empty(t, entries)
}
