package samples_test

import (
	"context"
	"errors"
	"io"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/ismelen/inkomi/back/manga-worker/internal/domain/models"
	"github.com/ismelen/inkomi/back/manga-worker/internal/infra/samples"
	"github.com/ismelen/inkomi/back/manga-worker/internal/usecases/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var filenames = map[models.PreviewSample]string{
	models.PreviewSampleColor: "color.jpg",
	models.PreviewSampleBW:    "bw.jpg",
}

func TestCloudSampleSource_Get_Repeated_DownloadsOnce(t *testing.T) {
	t.Parallel()

	// Arrange
	var downloads atomic.Int32
	cloud := &mocks.MockCloudStorage{DownloadFn: func(_ context.Context, id, filename string) (io.ReadCloser, error) {
		downloads.Add(1)
		return io.NopCloser(strings.NewReader(id + "/" + filename)), nil
	}}
	source := samples.NewCloudSampleSource(cloud, "samples", filenames, 1<<20)

	// Act
	var wg sync.WaitGroup
	for range 10 {
		wg.Go(func() {
			data, err := source.Get(context.Background(), models.PreviewSampleBW)
			assert.NoError(t, err)
			assert.Equal(t, "samples/bw.jpg", string(data))
		})
	}
	wg.Wait()

	// Assert
	assert.Equal(t, int32(1), downloads.Load())
}

func TestCloudSampleSource_Get_DownloadFails_RetriesNextTime(t *testing.T) {
	t.Parallel()

	// Arrange
	var calls atomic.Int32
	cloud := &mocks.MockCloudStorage{DownloadFn: func(context.Context, string, string) (io.ReadCloser, error) {
		if calls.Add(1) == 1 {
			return nil, errors.New("timeout")
		}
		return io.NopCloser(strings.NewReader("image")), nil
	}}
	source := samples.NewCloudSampleSource(cloud, "samples", filenames, 1<<20)

	// Act
	_, firstErr := source.Get(context.Background(), models.PreviewSampleColor)
	data, secondErr := source.Get(context.Background(), models.PreviewSampleColor)

	// Assert
	assert.Error(t, firstErr)
	require.NoError(t, secondErr)
	assert.Equal(t, "image", string(data))
}

func TestCloudSampleSource_Get_TooLarge_ReturnsErrSampleTooLarge(t *testing.T) {
	t.Parallel()

	// Arrange
	cloud := &mocks.MockCloudStorage{DownloadFn: func(context.Context, string, string) (io.ReadCloser, error) {
		return io.NopCloser(strings.NewReader("0123456789")), nil
	}}
	source := samples.NewCloudSampleSource(cloud, "samples", filenames, 5)

	// Act
	_, err := source.Get(context.Background(), models.PreviewSampleColor)

	// Assert
	assert.ErrorIs(t, err, models.ErrSampleTooLarge)
}

func TestCloudSampleSource_Get_UnknownSample_ReturnsErrUnknownSample(t *testing.T) {
	t.Parallel()

	// Arrange
	source := samples.NewCloudSampleSource(&mocks.MockCloudStorage{}, "samples", filenames, 5)

	// Act
	_, err := source.Get(context.Background(), models.PreviewSample("other"))

	// Assert
	assert.ErrorIs(t, err, models.ErrUnknownSample)
}
