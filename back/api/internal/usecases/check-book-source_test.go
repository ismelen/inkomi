package usecases

import (
	"context"
	"ismelen/inkomi/internal/test/mocks"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCheckUC(srv *httptest.Server, hasMirror bool, refreshResult bool) (*CheckBooksSourceUC, *mocks.BooksProviderMock) {
	url := ""
	if srv != nil {
		url = srv.URL
	}
	src := mocks.NewBooksSourceMock(url)
	prov := &mocks.BooksProviderMock{
		Source:        src,
		HasMirror:     hasMirror,
		RefreshResult: refreshResult,
	}
	return NewCheckBooksSourceUC(prov), prov
}

func TestCheckBooksSourceUC_Execute_WithMirrorAlive_ShouldReturnOk(t *testing.T) {
	t.Parallel()
	// Arrange
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	uc, prov := newCheckUC(srv, true, false)

	// Act
	ok, err := uc.Execute(context.Background())

	// Assert
	require.NoError(t, err)
	assert.True(t, ok)
	assert.False(t, prov.RefreshCalled)
}

func TestCheckBooksSourceUC_Execute_WithMirrorDead_RefreshCalled(t *testing.T) {
	t.Parallel()
	// Arrange
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	uc, prov := newCheckUC(srv, true, false)

	// Act
	_, err := uc.Execute(context.Background())

	// Assert
	require.NoError(t, err)
	assert.True(t, prov.RefreshCalled)
}

func TestCheckBooksSourceUC_Execute_MirrorDead_CooldownActive(t *testing.T) {
	t.Parallel()
	// Arrange
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	uc, prov := newCheckUC(srv, true, false)

	uc.mu.Lock()
	uc.lastCheck = time.Now()
	uc.mu.Unlock()

	// Act
	ok, err := uc.Execute(context.Background())

	// Assert
	require.NoError(t, err)
	assert.False(t, ok)
	assert.False(t, prov.RefreshCalled)
}

func TestCheckBooksSourceUC_Execute_NoMirror_RefreshSucceeds(t *testing.T) {
	t.Parallel()
	// Arrange
	uc, prov := newCheckUC(nil, false, true)

	// Act
	ok, err := uc.Execute(context.Background())

	// Assert
	require.NoError(t, err)
	assert.True(t, ok)
	assert.True(t, prov.RefreshCalled)
}

func TestCheckBooksSourceUC_Execute_NoMirror_RefreshFails(t *testing.T) {
	t.Parallel()
	// Arrange
	uc, prov := newCheckUC(nil, false, false)

	// Act
	ok, err := uc.Execute(context.Background())

	// Assert
	require.NoError(t, err)
	assert.False(t, ok)
	assert.True(t, prov.RefreshCalled)
}
