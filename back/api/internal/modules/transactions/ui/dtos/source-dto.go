package dtos

import "github.com/ismelen/inkomi/back/api/internal/modules/transactions/domain/models"

type NoItemsSourceDTO struct {
	Title            string                  `json:"title"`
	Type             models.SourceType       `json:"type"`
	Size             *int64                  `json:"size,omitempty"`
	Filename         string                  `json:"filename"`
	ConfigKey        *string                 `json:"config_hash,omitempty"`
	ReadingDirection models.ReadingDirection `json:"reading_direction"`
	ToCloud          bool                    `json:"to_cloud"`
	Kepubify         bool                    `json:"_"`
	Url              string                  `json:"url"`
	Id               string                  `json:"id"`
}

type SourceDTO struct {
	NoItemsSourceDTO
	ShouldJoin bool               `json:"should_join"`
	Items      []NoItemsSourceDTO `json:"items"`
}

func (d NoItemsSourceDTO) ToUploadItem() models.UploadItem {
	return models.UploadItem{
		Title:            d.Title,
		Type:             d.Type,
		Size:             d.Size,
		Filename:         d.Filename,
		ConfigKey:        d.ConfigKey,
		ReadingDirection: d.ReadingDirection,
		ToCloud:          d.ToCloud,
		Kepubify:         d.Kepubify,
		Url:              d.Url,
		Id:               d.Id,
	}
}

func (d SourceDTO) ToUploadSource() models.UploadSource {
	items := make([]models.UploadItem, len(d.Items))
	for i, item := range d.Items {
		items[i] = item.ToUploadItem()
	}
	return models.UploadSource{
		UploadItem: d.NoItemsSourceDTO.ToUploadItem(),
		ShouldJoin: d.ShouldJoin,
		Items:      items,
	}
}

func NewNoItemsSourceDTO(item models.UploadItem) NoItemsSourceDTO {
	return NoItemsSourceDTO{
		Title:            item.Title,
		Type:             item.Type,
		Size:             item.Size,
		Filename:         item.Filename,
		ConfigKey:        item.ConfigKey,
		ReadingDirection: item.ReadingDirection,
		ToCloud:          item.ToCloud,
		Kepubify:         item.Kepubify,
		Url:              item.Url,
		Id:               item.Id,
	}
}

func NewSourceDTO(source models.UploadSource) SourceDTO {
	items := make([]NoItemsSourceDTO, len(source.Items))
	for i, item := range source.Items {
		items[i] = NewNoItemsSourceDTO(item)
	}
	return SourceDTO{
		NoItemsSourceDTO: NewNoItemsSourceDTO(source.UploadItem),
		ShouldJoin:       source.ShouldJoin,
		Items:            items,
	}
}
