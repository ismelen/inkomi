package dtos

import (
	"encoding/json"
	"fmt"

	"github.com/ismelen/inkomi/back/api/internal/modules/transactions/domain/models"
)

type UploadRequestDTO struct {
	EReaderKey string               `json:"ereader_key"`
	Sources    []SourceDTO          `json:"sources"`
	Configs    map[string]ConfigDTO `json:"config"`
}

// ToUploadRequest serializes each config in the format the manga module reads.
func (d UploadRequestDTO) ToUploadRequest() (models.UploadRequest, error) {
	sources := make([]models.UploadSource, len(d.Sources))
	for i, source := range d.Sources {
		sources[i] = source.ToUploadSource()
	}
	configs := make(map[string]models.UploadConfig, len(d.Configs))
	for key, config := range d.Configs {
		data, err := json.Marshal(config)
		if err != nil {
			return models.UploadRequest{}, fmt.Errorf("config %s: %w", key, err)
		}
		configs[key] = models.UploadConfig{Data: string(data), Kepubify: config.Kepubify}
	}
	return models.UploadRequest{EReaderKey: d.EReaderKey, Sources: sources, Configs: configs}, nil
}
