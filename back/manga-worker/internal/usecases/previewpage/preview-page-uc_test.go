package previewpage_test

import (
	"context"
	"errors"
	"testing"

	"github.com/ismelen/inkomi/back/manga-worker/internal/domain/models"
	"github.com/ismelen/inkomi/back/manga-worker/internal/usecases/mocks"
	"github.com/ismelen/inkomi/back/manga-worker/internal/usecases/previewpage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPreviewPageUC_Execute_Valid_ProcessesSampleAsSinglePage(t *testing.T) {
	t.Parallel()

	// Arrange
	var gotSample models.PreviewSample
	source := &mocks.MockSampleSource{GetFn: func(_ context.Context, sample models.PreviewSample) ([]byte, error) {
		gotSample = sample
		return []byte("original"), nil
	}}
	processor := &mocks.MockPageProcessor{
		ProcessSingleFn: func(_ context.Context, data []byte, cfg models.MangaConfig) (models.ProcessedPage, error) {
			assert.Equal(t, "original", string(data))
			assert.InDelta(t, 1.4, cfg.Gamma, 1e-9)
			return models.ProcessedPage{Data: []byte("processed"), Format: models.ImageFormatPNG}, nil
		},
		ProcessFn: func(context.Context, []byte, models.MangaConfig, bool) ([]models.ProcessedPage, error) {
			t.Fatal("the preview must never split pages")
			return nil, nil
		},
	}
	cfg := models.DefaultMangaConfig()
	cfg.Gamma = 1.4

	// Act
	page, err := previewpage.NewPreviewPageUC(source, processor).Execute(context.Background(), models.PreviewSampleColor, cfg)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, models.PreviewSampleColor, gotSample)
	assert.Equal(t, "processed", string(page.Data))
	assert.Equal(t, models.ImageFormatPNG, page.Format)
}

func TestPreviewPageUC_Execute_TooLarge_ReturnsErrorWithoutProcessing(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name          string
		width, height int
	}{
		{"width", models.PreviewMaxSize + 1, 100},
		{"height", 100, models.PreviewMaxSize + 1},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			called := false
			source := &mocks.MockSampleSource{GetFn: func(context.Context, models.PreviewSample) ([]byte, error) {
				called = true
				return nil, nil
			}}
			cfg := models.DefaultMangaConfig()
			cfg.TargetWidth, cfg.TargetHeight = c.width, c.height

			// Act
			_, err := previewpage.NewPreviewPageUC(source, &mocks.MockPageProcessor{}).Execute(context.Background(), models.PreviewSampleBW, cfg)

			// Assert
			assert.ErrorIs(t, err, models.ErrPreviewTooLarge)
			assert.False(t, called)
		})
	}
}

func TestPreviewPageUC_Execute_Failure_ReturnsWrappedError(t *testing.T) {
	t.Parallel()

	errBoom := errors.New("boom")
	cases := []struct {
		name      string
		source    *mocks.MockSampleSource
		processor *mocks.MockPageProcessor
	}{
		{
			"sample",
			&mocks.MockSampleSource{GetFn: func(context.Context, models.PreviewSample) ([]byte, error) { return nil, errBoom }},
			&mocks.MockPageProcessor{},
		},
		{
			"processing",
			&mocks.MockSampleSource{},
			&mocks.MockPageProcessor{ProcessSingleFn: func(context.Context, []byte, models.MangaConfig) (models.ProcessedPage, error) {
				return models.ProcessedPage{}, errBoom
			}},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			// Act
			_, err := previewpage.NewPreviewPageUC(c.source, c.processor).Execute(context.Background(), models.PreviewSampleBW, models.DefaultMangaConfig())

			// Assert
			assert.ErrorIs(t, err, errBoom)
		})
	}
}
