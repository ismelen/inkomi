package dtos

import "github.com/ismelen/inkomi/back/transaction-manager/internal/domain/models"

type NoItemsSourceDTO struct {
	Title            string                  `json:"title"`
	Type             models.SourceType       `json:"type"`
	Size             *int64                  `json:"size,omitempty"`
	Filename         string                  `json:"filename"`
	ConfigHash       *string                 `json:"config_hash,omitempty"`
	ReadingDirection models.ReadingDirection `json:"reading_direction"`
	Kepubify         bool                    `json:"_"`
	Url              string                  `json:"url"`
	Id               string                  `json:"id"`
}

type SourceDTO struct {
	NoItemsSourceDTO
	ShouldJoin bool               `json:"should_join"`
	Items      []NoItemsSourceDTO `json:"items"`
}
