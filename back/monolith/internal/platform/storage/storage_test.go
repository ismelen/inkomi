package storage_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/ismelen/inkomi/back/monolith/internal/platform/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newStore(t *testing.T) (*storage.LocalStorage, string) {
	t.Helper()
	dir := t.TempDir()
	store, err := storage.NewLocalStorage(filepath.Join(dir, "files"), filepath.Join(dir, "tmp"))
	require.NoError(t, err)
	return store, dir
}

func TestLocalStorage_UploadAndDownloadFile_RoundTrip(t *testing.T) {
	t.Parallel()

	// Arrange
	store, dir := newStore(t)
	ctx := context.Background()
	src := filepath.Join(dir, "tmp", "book.epub")
	require.NoError(t, os.WriteFile(src, []byte("epub-bytes"), 0o600))

	// Act
	size, err := store.UploadFile(ctx, "id-1", "Book.epub", src)
	require.NoError(t, err)
	require.NoError(t, os.Remove(src))
	dst := filepath.Join(dir, "tmp", "copy.epub")
	downloadErr := store.DownloadFile(ctx, "id-1", "Book.epub", dst)

	// Assert
	require.NoError(t, downloadErr)
	assert.Equal(t, int64(10), size)
	data, err := os.ReadFile(dst)
	require.NoError(t, err)
	assert.Equal(t, "epub-bytes", string(data))
	stored, err := store.Size(ctx, "id-1", "Book.epub")
	require.NoError(t, err)
	assert.Equal(t, int64(10), stored)
}

func TestLocalStorage_Delete_RemovesFileAndEmptyDir(t *testing.T) {
	t.Parallel()

	// Arrange
	store, dir := newStore(t)
	ctx := context.Background()
	require.NoError(t, store.Upload(ctx, "id-1", "a.cbz", strings.NewReader("a")))

	// Act
	err := store.Delete(ctx, "id-1", "a.cbz")
	again := store.Delete(ctx, "id-1", "a.cbz")

	// Assert
	require.NoError(t, err)
	require.NoError(t, again)
	assert.False(t, store.Exists("id-1", "a.cbz"))
	_, statErr := os.Stat(filepath.Join(dir, "files", "id-1"))
	assert.True(t, os.IsNotExist(statErr))
}

func TestLocalStorage_Path_RejectsTraversal(t *testing.T) {
	t.Parallel()

	store, _ := newStore(t)
	for _, key := range [][2]string{{"..", "x"}, {"id", "../x"}, {"id", "a/b"}, {"id", `a\b`}, {"", "x"}} {
		_, err := store.Path(key[0], key[1])
		assert.ErrorIs(t, err, storage.ErrInvalidKey, key)
	}
}

func TestLocalStorage_Write_OverLimit_StoresNothing(t *testing.T) {
	t.Parallel()

	// Arrange
	store, _ := newStore(t)

	// Act
	_, err := store.Write("id-1", "big.cbz", strings.NewReader("0123456789"), 5)

	// Assert
	assert.ErrorIs(t, err, storage.ErrTooLarge)
	assert.False(t, store.Exists("id-1", "big.cbz"))
}

func TestLocalStorage_Usage_SumsIds(t *testing.T) {
	t.Parallel()

	// Arrange
	store, _ := newStore(t)
	ctx := context.Background()
	require.NoError(t, store.Upload(ctx, "a", "1", strings.NewReader("12345")))
	require.NoError(t, store.Upload(ctx, "a", "2", strings.NewReader("123")))
	require.NoError(t, store.Upload(ctx, "b", "1", strings.NewReader("1")))

	// Act & Assert
	assert.Equal(t, int64(8), store.Usage([]string{"a", "missing"}))
	assert.Equal(t, int64(9), store.Usage([]string{"a", "b"}))
}

func TestURLSigner_Verify(t *testing.T) {
	t.Parallel()

	now := time.Unix(1_800_000_000, 0)
	signer := storage.NewURLSigner([]byte("key"), "https://x.test", func() time.Time { return now })
	raw, _ := signer.URL(http.MethodPut, "id-1", "Capítulo 1.cbz", 15*time.Minute, 100)
	u, err := url.Parse(raw)
	require.NoError(t, err)
	q := u.Query()

	max, err := signer.Verify(http.MethodPut, "id-1", "Capítulo 1.cbz", q)
	require.NoError(t, err)
	assert.Equal(t, int64(100), max)

	_, err = signer.Verify(http.MethodGet, "id-1", "Capítulo 1.cbz", q)
	assert.ErrorIs(t, err, storage.ErrInvalidSignature, "other method")
	_, err = signer.Verify(http.MethodPut, "id-2", "Capítulo 1.cbz", q)
	assert.ErrorIs(t, err, storage.ErrInvalidSignature, "other file")
	tampered := url.Values{"exp": q["exp"], "max": {"999"}, "sig": q["sig"]}
	_, err = signer.Verify(http.MethodPut, "id-1", "Capítulo 1.cbz", tampered)
	assert.ErrorIs(t, err, storage.ErrInvalidSignature, "bigger limit")
	later := storage.NewURLSigner([]byte("key"), "https://x.test", func() time.Time { return now.Add(time.Hour) })
	_, err = later.Verify(http.MethodPut, "id-1", "Capítulo 1.cbz", q)
	assert.ErrorIs(t, err, storage.ErrInvalidSignature, "expired")
}

func TestFilesHandler_PutAndGet(t *testing.T) {
	t.Parallel()

	// Arrange
	store, _ := newStore(t)
	signer := storage.NewURLSigner([]byte("key"), "", time.Now)
	router := chi.NewRouter()
	storage.NewFilesHandler(store, signer).Setup(router)
	srv := httptest.NewServer(router)
	defer srv.Close()
	put, _ := signer.URL(http.MethodPut, "id-1", "Capítulo 1.cbz", time.Minute, 5)
	get, _ := signer.URL(http.MethodGet, "id-1", "Capítulo 1.cbz", time.Minute, -1)
	do := func(method, path string, body string) *http.Response {
		req, err := http.NewRequest(method, srv.URL+path, strings.NewReader(body))
		require.NoError(t, err)
		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		t.Cleanup(func() { _ = res.Body.Close() })
		return res
	}

	// Act & Assert
	assert.Equal(t, http.StatusRequestEntityTooLarge, do(http.MethodPut, put, "123456").StatusCode)
	assert.Equal(t, http.StatusForbidden, do(http.MethodPut, get, "12345").StatusCode)
	assert.Equal(t, http.StatusNotFound, do(http.MethodGet, get, "").StatusCode)
	assert.Equal(t, http.StatusOK, do(http.MethodPut, put, "12345").StatusCode)

	res := do(http.MethodGet, get, "")
	require.Equal(t, http.StatusOK, res.StatusCode)
	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	assert.Equal(t, "12345", string(body))
	assert.Equal(t, "attachment; filename*=UTF-8''Cap%C3%ADtulo%201.cbz", res.Header.Get("Content-Disposition"))
}

func TestJanitor_Sweep_RemovesOnlyExpired(t *testing.T) {
	t.Parallel()

	// Arrange
	store, dir := newStore(t)
	ctx := context.Background()
	require.NoError(t, store.Upload(ctx, "old", "a", strings.NewReader("a")))
	require.NoError(t, store.Upload(ctx, "new", "a", strings.NewReader("a")))
	past := time.Now().Add(-10 * 24 * time.Hour)
	for _, p := range []string{filepath.Join(dir, "files", "old", "a"), filepath.Join(dir, "files", "old")} {
		require.NoError(t, os.Chtimes(p, past, past))
	}

	// Act
	removed := storage.NewJanitor(store, 7*24*time.Hour, time.Now).Sweep()

	// Assert
	assert.Equal(t, 1, removed)
	assert.False(t, store.Exists("old", "a"))
	assert.True(t, store.Exists("new", "a"))
}
