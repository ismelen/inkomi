package dtos_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ismelen/inkomi/back/api/internal/modules/transactions/domain/models"
	"github.com/ismelen/inkomi/back/api/internal/modules/transactions/ui/dtos"
)

func TestUploadRequestDTO_ToUploadRequest_ValidRequest_MapsSourcesAndSerializesConfigs(t *testing.T) {
	t.Parallel()

	// Arrange
	var dto dtos.UploadRequestDTO
	require.NoError(t, json.Unmarshal([]byte(`{
		"ereader_key": "K1",
		"config": {"c1": {"ereaderKey": "KoMT", "grayscale": true}},
		"sources": [{"title": "Folder", "type": "folder", "should_join": true, "config_hash": "c1",
			"items": [{"title": "Ch 1", "type": "file", "filename": "1.cbz"}]}]
	}`), &dto))
	expectedData, err := json.Marshal(dto.Configs["c1"])
	require.NoError(t, err)

	// Act
	request, err := dto.ToUploadRequest()

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "K1", request.EReaderKey)
	assert.Equal(t, models.UploadConfig{Data: string(expectedData), Kepubify: true}, request.Configs["c1"])
	require.Len(t, request.Sources, 1)
	assert.Equal(t, "Folder", request.Sources[0].Title)
	assert.True(t, request.Sources[0].ShouldJoin)
	assert.Equal(t, "c1", *request.Sources[0].ConfigKey)
	require.Len(t, request.Sources[0].Items, 1)
	assert.Equal(t, "1.cbz", request.Sources[0].Items[0].Filename)
}

func TestNewSourceDTO_SavedSource_KeepsIdsAndUrls(t *testing.T) {
	t.Parallel()

	// Arrange
	source := models.UploadSource{
		UploadItem: models.UploadItem{Id: "root", Type: models.SourceTypeFolder},
		ShouldJoin: true,
		Items:      []models.UploadItem{{Id: "child", Type: models.SourceTypeFile, Url: "https://u"}},
	}

	// Act
	dto := dtos.NewSourceDTO(source)

	// Assert
	assert.Equal(t, "root", dto.Id)
	assert.True(t, dto.ShouldJoin)
	require.Len(t, dto.Items, 1)
	assert.Equal(t, "child", dto.Items[0].Id)
	assert.Equal(t, "https://u", dto.Items[0].Url)
}
