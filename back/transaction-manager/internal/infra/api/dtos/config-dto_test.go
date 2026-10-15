package dtos_test

import (
	"encoding/json"
	"testing"

	"github.com/ismelen/inkomi/back/transaction-manager/internal/infra/api/dtos"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigDTO_UnmarshalJSON_MissingFields_AppliesDefaults(t *testing.T) {
	t.Parallel()

	// Arrange
	var cfg dtos.ConfigDTO

	// Act
	err := json.Unmarshal([]byte(`{"ereaderKey":"K1","sharpen":{"enabled":true}}`), &cfg)

	// Assert
	require.NoError(t, err)
	assert.InDelta(t, -1, cfg.CropTolerance, 1e-6)
	assert.InDelta(t, 1, cfg.Gamma, 1e-6)
	assert.Equal(t, int8(90), cfg.Quality)
	assert.True(t, cfg.Sharpen.Enabled)
	assert.Equal(t, int8(1), cfg.Sharpen.Radius)
	assert.InDelta(t, 3, cfg.Sharpen.M2, 1e-6)
}

func TestConfigDTO_MarshalJSON_Valid_EmitsCamelCaseWithTargetSize(t *testing.T) {
	t.Parallel()

	// Arrange
	var cfg dtos.ConfigDTO
	require.NoError(t, json.Unmarshal([]byte(`{"ereaderKey":"K1","palette8bit":true}`), &cfg))
	cfg.Hash = "hash"

	// Act
	data, err := json.Marshal(cfg)

	// Assert
	require.NoError(t, err)
	var out map[string]any
	require.NoError(t, json.Unmarshal(data, &out))
	assert.Equal(t, true, out["palette8bit"])
	assert.Equal(t, float64(600), out["targetWidth"])
	assert.Equal(t, float64(670), out["targetHeight"])
	assert.Contains(t, out, "blackPointThreshold")
	assert.NotContains(t, out, "Hash")
	assert.NotContains(t, out, "Kepubify")
}

func TestConfigDTO_UnmarshalJSON_InvalidEReader_ReturnsError(t *testing.T) {
	t.Parallel()

	// Arrange
	var cfg dtos.ConfigDTO

	// Act
	err := json.Unmarshal([]byte(`{"ereaderKey":"NOPE"}`), &cfg)

	// Assert
	assert.Error(t, err)
}
