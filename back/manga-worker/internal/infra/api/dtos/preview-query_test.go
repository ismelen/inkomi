package dtos_test

import (
	"net/url"
	"testing"

	"github.com/ismelen/inkomi/back/manga-worker/internal/domain/models"
	"github.com/ismelen/inkomi/back/manga-worker/internal/infra/api/dtos"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func field(t *testing.T, form dtos.PreviewForm, name string) dtos.PreviewFormField {
	t.Helper()

	for _, f := range form.Fields {
		if f.Name == name {
			return f
		}
	}
	require.Failf(t, "missing field", "%s", name)
	return dtos.PreviewFormField{}
}

func TestParsePreviewQuery_Empty_ReturnsDefaultsAndBWSample(t *testing.T) {
	t.Parallel()

	// Act
	req, err := dtos.ParsePreviewQuery(url.Values{})

	// Assert
	require.NoError(t, err)
	assert.Equal(t, models.PreviewSampleBW, req.Sample)
	assert.Equal(t, models.DefaultMangaConfig(), req.Config)
	assert.Equal(t, "1072", field(t, req.Form, "targetWidth").Value)
	assert.Equal(t, "0.06", field(t, req.Form, "blackPointThreshold").Value)
	assert.False(t, field(t, req.Form, "grayscale").Checked)
}

func TestParsePreviewQuery_Submitted_ParsesValuesAndCheckboxes(t *testing.T) {
	t.Parallel()

	// Arrange
	values := url.Values{
		"sample":         {"color"},
		"submitted":      {"1"},
		"gamma":          {" 1.8 "},
		"quality":        {"75"},
		"dithering":      {"1"},
		"sharpenEnabled": {"1"},
		"sharpenRadius":  {"2"},
		"targetWidth":    {""},
	}

	// Act
	req, err := dtos.ParsePreviewQuery(values)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, models.PreviewSampleColor, req.Sample)
	assert.InDelta(t, 1.8, req.Config.Gamma, 1e-9)
	assert.Equal(t, 75, req.Config.Quality)
	assert.True(t, req.Config.Dithering)
	assert.False(t, req.Config.Grayscale)
	assert.True(t, req.Config.Sharpen.Enabled)
	assert.Equal(t, 2, req.Config.Sharpen.Radius)
	assert.InDelta(t, 1.5, req.Config.Sharpen.X1, 1e-9)
	assert.Equal(t, 1072, req.Config.TargetWidth)
	assert.Equal(t, "1.8", field(t, req.Form, "gamma").Value)
	assert.True(t, field(t, req.Form, "dithering").Checked)
}

func TestParsePreviewQuery_NotSubmitted_IgnoresCheckboxes(t *testing.T) {
	t.Parallel()

	// Act
	req, err := dtos.ParsePreviewQuery(url.Values{"dithering": {"1"}})

	// Assert
	require.NoError(t, err)
	assert.False(t, req.Config.Dithering)
}

func TestParsePreviewQuery_Invalid_ReturnsErrorAndForm(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		values url.Values
		want   error
	}{
		{"unknown sample", url.Values{"sample": {"../secret"}}, models.ErrUnknownSample},
		{"not a number", url.Values{"gamma": {"abc"}}, models.ErrInvalidMangaConfig},
		{"nan", url.Values{"gamma": {"NaN"}}, models.ErrInvalidMangaConfig},
		{"out of range", url.Values{"gamma": {"9"}}, models.ErrInvalidMangaConfig},
		{"decimal integer", url.Values{"quality": {"7.5"}}, models.ErrInvalidMangaConfig},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			// Act
			req, err := dtos.ParsePreviewQuery(c.values)

			// Assert
			assert.ErrorIs(t, err, c.want)
			require.NotNil(t, req)
			assert.NotEmpty(t, req.Form.Fields)
		})
	}
}

func TestPreviewForm_Query_RoundTrip_KeepsConfig(t *testing.T) {
	t.Parallel()

	// Arrange
	values := url.Values{
		"sample":      {"color"},
		"submitted":   {"1"},
		"contrast":    {"1.4"},
		"grayscale":   {"1"},
		"palette8bit": {"1"},
	}
	first, err := dtos.ParsePreviewQuery(values)
	require.NoError(t, err)

	// Act
	encoded, err := url.ParseQuery(first.Form.Query())
	require.NoError(t, err)
	second, err := dtos.ParsePreviewQuery(encoded)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, first.Sample, second.Sample)
	assert.Equal(t, first.Config, second.Config)
}
