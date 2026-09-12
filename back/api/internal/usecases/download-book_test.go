package usecases

import (
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ismelen/inkomi/internal/domain/book"
	"ismelen/inkomi/internal/test/mocks"
)

func makeDownload() *book.LibgenDownload {
	return &book.LibgenDownload{
		Filename: "test.epub",
		Stream:   io.NopCloser(strings.NewReader("epub content")),
	}
}

func newDownloadUC(hasMirror bool, src *mocks.BooksSourceMock) (*DownloadBookUC, *mocks.BooksProviderMock) {
	prov := &mocks.BooksProviderMock{
		Source:    src,
		HasMirror: hasMirror,
	}
	return NewDownloadBookUC(prov), prov
}

func TestDownloadBookUC_Execute_ZeroRetries_Error(t *testing.T) {
	t.Parallel()
	// Arrange
	src := mocks.NewBooksSourceMock("http://test.example")
	uc, _ := newDownloadUC(true, src)

	// Act
	result, err := uc.Execute("abc123", 0)

	// Assert
	require.Error(t, err)
	assert.Nil(t, result)
}

func TestDownloadBookUC_Execute_NoMirror_Error(t *testing.T) {
	t.Parallel()
	// Arrange
	src := mocks.NewBooksSourceMock("http://test.example")
	uc, _ := newDownloadUC(false, src)

	// Act
	_, err := uc.Execute("abc123", 3)

	// Assert
	require.Error(t, err)
	assert.Contains(t, strings.ToLower(err.Error()), "mirror")
}

func TestDownloadBookUC_Execute_Normal_Success(t *testing.T) {
	t.Parallel()
	// Arrange
	src := mocks.NewBooksSourceMock("http://test.example")
	src.DownloadResult = makeDownload()
	uc, _ := newDownloadUC(true, src)

	// Act
	result, err := uc.Execute("abc123", 3)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "test.epub", result.Filename)
}

func TestDownloadBookUC_Execute_RetriesAndSucceeds_ShouldSucceed(t *testing.T) {
	t.Parallel()
	// Arrange
	downloadErr := errors.New("temporary failure")
	src := mocks.NewBooksSourceMock("http://test.example")
	src.DownloadErr = downloadErr
	src.DownloadResult = makeDownload()
	src.WithDownloadFailN(2)

	uc, prov := newDownloadUC(true, src)
	prov.RefreshResult = true // Refresh devuelve updated=true para que el retry continúe

	// Act
	result, err := uc.Execute("abc123", 3)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestDownloadBookUC_Execute_NoWorkingMirrors_ShouldError(t *testing.T) {
	t.Parallel()
	// Arrange
	downloadErr := errors.New("always fails")
	src := mocks.NewBooksSourceMock("http://test.example")
	src.DownloadErr = downloadErr
	src.WithDownloadFailN(999) // always fail

	uc, prov := newDownloadUC(true, src)
	prov.RefreshResult = false // Refresh devuelve updated=false → "no working mirrors"

	// Act
	result, err := uc.Execute("abc123", 2)

	// Assert
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "no working mirrors")
}

func TestDownloadBookUC_Execute_ExhaustsRetries_ShouldError(t *testing.T) {
	t.Parallel()
	// Arrange
	downloadErr := errors.New("always fails")
	src := mocks.NewBooksSourceMock("http://test.example")
	src.DownloadErr = downloadErr
	src.WithDownloadFailN(999) // always fail

	uc, prov := newDownloadUC(true, src)
	prov.RefreshResult = true // Refresh ok, pero se agotan los reintentos

	// Act
	result, err := uc.Execute("abc123", 2)

	// Assert
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "no working mirror")
}
