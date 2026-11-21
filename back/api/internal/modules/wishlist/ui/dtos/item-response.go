package dtos

import (
	"time"

	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/domain/models"
)

type ItemResponse struct {
	Id        int64     `json:"id"`
	Kind      string    `json:"kind"`
	Title     string    `json:"title"`
	Author    string    `json:"author"`
	WorkKey   string    `json:"workKey"`
	AuthorKey string    `json:"authorKey"`
	CoverURL  string    `json:"coverUrl"`
	CreatedAt time.Time `json:"createdAt"`
}

func NewItemResponse(item models.Item) ItemResponse {
	return ItemResponse{
		Id:        item.Id,
		Kind:      string(item.Kind),
		Title:     item.Title,
		Author:    item.Author,
		WorkKey:   item.WorkKey,
		AuthorKey: item.AuthorKey,
		CoverURL:  item.CoverURL,
		CreatedAt: item.CreatedAt,
	}
}
