package dtos

import "github.com/ismelen/inkomi/back/api/internal/modules/wishlist/domain/models"

type SuggestionResponse struct {
	Kind      string `json:"kind"`
	Title     string `json:"title"`
	Author    string `json:"author"`
	WorkKey   string `json:"workKey"`
	AuthorKey string `json:"authorKey"`
	Year      int    `json:"year"`
	CoverURL  string `json:"coverUrl"`
}

func NewSuggestionResponse(s models.Suggestion) SuggestionResponse {
	return SuggestionResponse{
		Kind:      string(s.Kind),
		Title:     s.Title,
		Author:    s.Author,
		WorkKey:   s.WorkKey,
		AuthorKey: s.AuthorKey,
		Year:      s.Year,
		CoverURL:  s.CoverURL,
	}
}
