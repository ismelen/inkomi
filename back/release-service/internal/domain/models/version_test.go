package models_test

import (
	"testing"

	"github.com/ismelen/inkomi/back/release-service/internal/domain/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseVersion_ValidTag_ReturnsVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		raw  string
		want models.Version
	}{
		{"v1.3.7", models.Version{Major: 1, Minor: 3, Patch: 7}},
		{"2.0.0", models.Version{Major: 2}},
		{" v10.20.30 ", models.Version{Major: 10, Minor: 20, Patch: 30}},
	}

	for _, tt := range tests {
		t.Run(tt.raw, func(t *testing.T) {
			t.Parallel()

			// Act
			got, err := models.ParseVersion(tt.raw)

			// Assert
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.want.String(), got.String())
		})
	}
}

func TestParseVersion_InvalidTag_ReturnsError(t *testing.T) {
	t.Parallel()

	for _, raw := range []string{"", "v1.3", "1.3.7.1", "v1.x.0", "1.-1.0", "latest"} {
		t.Run(raw, func(t *testing.T) {
			t.Parallel()

			// Act
			_, err := models.ParseVersion(raw)

			// Assert
			assert.ErrorIs(t, err, models.ErrInvalidVersion)
		})
	}
}
