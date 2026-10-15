package models_test

import (
	"encoding/json"
	"testing"

	"github.com/ismelen/inkomi/back/manga-worker/internal/domain/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMangaConfig_Empty_ReturnsDefaults(t *testing.T) {
	t.Parallel()

	for _, raw := range []json.RawMessage{nil, json.RawMessage(`{}`)} {
		// Act
		cfg, err := models.NewMangaConfig(raw)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, models.DefaultMangaConfig(), cfg)
		assert.False(t, cfg.CropEnabled())
	}
}

func TestNewMangaConfig_Partial_KeepsDefaultsForMissingFields(t *testing.T) {
	t.Parallel()

	// Arrange
	raw := json.RawMessage(`{"gamma":1.8,"cropTolerance":0.1,"palette8bit":true,"sharpen":{"enabled":true,"radius":2}}`)

	// Act
	cfg, err := models.NewMangaConfig(raw)

	// Assert
	require.NoError(t, err)
	assert.InDelta(t, 1.8, cfg.Gamma, 1e-9)
	assert.True(t, cfg.CropEnabled())
	assert.True(t, cfg.Palette8bit)
	assert.True(t, cfg.Sharpen.Enabled)
	assert.Equal(t, 2, cfg.Sharpen.Radius)
	assert.InDelta(t, 1.5, cfg.Sharpen.X1, 1e-9)
	assert.Equal(t, 1072, cfg.TargetWidth)
	assert.Equal(t, 90, cfg.Quality)
}

func TestNewMangaConfig_Invalid_ReturnsErrInvalidMangaConfig(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		raw  string
	}{
		{"malformed", `{"gamma":`},
		{"gamma too low", `{"gamma":0}`},
		{"brightness too high", `{"brightness":1.5}`},
		{"contrast negative", `{"contrast":-0.1}`},
		{"crop below -1", `{"cropTolerance":-0.5}`},
		{"crop above 1", `{"cropTolerance":1.5}`},
		{"quality zero", `{"quality":0}`},
		{"negative width", `{"targetWidth":-1}`},
		{"black above white", `{"blackPointThreshold":0.9,"whitePointThreshold":0.5}`},
		{"enabled sharpen radius", `{"sharpen":{"enabled":true,"radius":9}}`},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			// Act
			_, err := models.NewMangaConfig(json.RawMessage(c.raw))

			// Assert
			assert.ErrorIs(t, err, models.ErrInvalidMangaConfig)
		})
	}
}

func TestNewMangaConfig_DisabledSharpenOutOfRange_IsIgnored(t *testing.T) {
	t.Parallel()

	// Act
	_, err := models.NewMangaConfig(json.RawMessage(`{"sharpen":{"enabled":false,"radius":9}}`))

	// Assert
	assert.NoError(t, err)
}

func TestSourceFormatFromFilename(t *testing.T) {
	t.Parallel()

	cases := []struct {
		filename string
		want     models.SourceFormat
		err      error
	}{
		{"chapter.cbz", models.SourceFormatCBZ, nil},
		{"chapter.ZIP", models.SourceFormatCBZ, nil},
		{"chapter.pdf", models.SourceFormatPDF, nil},
		{"chapter.epub", "", models.ErrUnsupportedFormat},
		{"chapter", "", models.ErrUnsupportedFormat},
	}

	for _, c := range cases {
		t.Run(c.filename, func(t *testing.T) {
			t.Parallel()

			// Act
			got, err := models.SourceFormatFromFilename(c.filename)

			// Assert
			assert.ErrorIs(t, err, c.err)
			assert.Equal(t, c.want, got)
		})
	}
}

func TestImageFormat_ExtAndMediaType(t *testing.T) {
	t.Parallel()

	// Assert
	assert.Equal(t, ".jpg", models.ImageFormatJPEG.Ext())
	assert.Equal(t, "image/jpeg", models.ImageFormatJPEG.MediaType())
	assert.Equal(t, ".png", models.ImageFormatPNG.Ext())
	assert.Equal(t, "image/png", models.ImageFormatPNG.MediaType())
}
