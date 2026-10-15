package handlers_test

import (
	"context"
	"errors"
	"html"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ismelen/inkomi/back/manga-worker/internal/domain/models"
	"github.com/ismelen/inkomi/back/manga-worker/internal/infra/api/handlers"
	"github.com/ismelen/inkomi/back/manga-worker/internal/usecases/getsample"
	"github.com/ismelen/inkomi/back/manga-worker/internal/usecases/mocks"
	"github.com/ismelen/inkomi/back/manga-worker/internal/usecases/previewpage"
	"github.com/stretchr/testify/assert"
)

func newHandler(source *mocks.MockSampleSource, processor *mocks.MockPageProcessor) *handlers.PreviewHandler {
	return handlers.NewPreviewHandler(previewpage.NewPreviewPageUC(source, processor), getsample.NewGetSampleUC(source))
}

func serve(handler http.HandlerFunc, target string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	handler(rec, httptest.NewRequest(http.MethodGet, target, nil))
	return rec
}

func TestPreviewHandler_HandlePage_ValidQuery_RendersBothImagesAndForm(t *testing.T) {
	t.Parallel()

	// Arrange
	h := newHandler(&mocks.MockSampleSource{}, &mocks.MockPageProcessor{})

	// Act
	rec := serve(h.HandlePage, "/preview?sample=color&submitted=1&gamma=1.3&dithering=1")

	// Assert
	body := html.UnescapeString(rec.Body.String())
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Header().Get("Content-Type"), "text/html")
	assert.Contains(t, body, `src="/preview/original?sample=color"`)
	assert.Contains(t, body, `src="/preview/processed?`)
	assert.Contains(t, body, "gamma=1.3")
	assert.Contains(t, body, `name="gamma" value="1.3"`)
	assert.Contains(t, body, `name="dithering" value="1" checked`)
	assert.Contains(t, body, `<option value="color" selected>`)
	assert.NotContains(t, body, `class="error"`)
}

func TestPreviewHandler_HandlePage_InvalidQuery_ShowsErrorWithoutProcessedImage(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		target string
	}{
		{"invalid config", "/preview?sample=bw&gamma=9"},
		{"too large", "/preview?sample=bw&targetWidth=5000"},
		{"unknown sample", "/preview?sample=<script>"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			h := newHandler(&mocks.MockSampleSource{}, &mocks.MockPageProcessor{})

			// Act
			rec := serve(h.HandlePage, c.target)

			// Assert
			body := rec.Body.String()
			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Contains(t, body, `class="error"`)
			assert.NotContains(t, body, "/preview/processed")
			assert.NotContains(t, body, "<script>")
			assert.Contains(t, body, `src="/preview/original?sample=bw"`)
		})
	}
}

func TestPreviewHandler_HandleOriginal_Sample_ReturnsCachedImage(t *testing.T) {
	t.Parallel()

	// Arrange
	png := "\x89PNG\r\n\x1a\nrest"
	h := newHandler(&mocks.MockSampleSource{GetFn: func(context.Context, models.PreviewSample) ([]byte, error) {
		return []byte(png), nil
	}}, &mocks.MockPageProcessor{})

	// Act
	rec := serve(h.HandleOriginal, "/preview/original?sample=bw")

	// Assert
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "image/png", rec.Header().Get("Content-Type"))
	assert.Equal(t, "public, max-age=86400", rec.Header().Get("Cache-Control"))
	assert.Equal(t, png, rec.Body.String())
}

func TestPreviewHandler_HandleOriginal_UnknownSample_Returns400(t *testing.T) {
	t.Parallel()

	// Arrange
	h := newHandler(&mocks.MockSampleSource{}, &mocks.MockPageProcessor{})

	// Act
	rec := serve(h.HandleOriginal, "/preview/original?sample=x")

	// Assert
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestPreviewHandler_HandleProcessed_Valid_ReturnsImageWithFormat(t *testing.T) {
	t.Parallel()

	// Arrange
	processor := &mocks.MockPageProcessor{ProcessSingleFn: func(_ context.Context, _ []byte, cfg models.MangaConfig) (models.ProcessedPage, error) {
		format := models.ImageFormatJPEG
		if cfg.Palette8bit {
			format = models.ImageFormatPNG
		}
		return models.ProcessedPage{Data: []byte("img"), Format: format}, nil
	}}
	h := newHandler(&mocks.MockSampleSource{}, processor)

	// Act
	rec := serve(h.HandleProcessed, "/preview/processed?sample=bw&submitted=1&palette8bit=1")

	// Assert
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "image/png", rec.Header().Get("Content-Type"))
	assert.Equal(t, "public, max-age=86400", rec.Header().Get("Cache-Control"))
	assert.Equal(t, "img", rec.Body.String())
}

func TestPreviewHandler_HandleProcessed_Errors_MapToStatus(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		target string
		err    error
		want   int
	}{
		{"invalid config", "/preview/processed?sample=bw&gamma=9", nil, http.StatusBadRequest},
		{"too large", "/preview/processed?sample=bw&targetHeight=3000", nil, http.StatusBadRequest},
		{"timeout", "/preview/processed?sample=bw", context.DeadlineExceeded, http.StatusGatewayTimeout},
		{"internal", "/preview/processed?sample=bw", errors.New("vips crashed"), http.StatusInternalServerError},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			processor := &mocks.MockPageProcessor{ProcessSingleFn: func(context.Context, []byte, models.MangaConfig) (models.ProcessedPage, error) {
				return models.ProcessedPage{}, c.err
			}}
			h := newHandler(&mocks.MockSampleSource{}, processor)

			// Act
			rec := serve(h.HandleProcessed, c.target)

			// Assert
			assert.Equal(t, c.want, rec.Code)
			assert.False(t, strings.HasPrefix(rec.Header().Get("Content-Type"), "image/"))
		})
	}
}
