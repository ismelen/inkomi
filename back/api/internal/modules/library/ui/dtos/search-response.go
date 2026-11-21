package dtos

import "github.com/ismelen/inkomi/back/api/internal/modules/library/domain/models"

type SearchResponse struct {
	Books []BookResponse `json:"books"`
	// NextPage is the page to ask for to see more books, or 0 when there are no more.
	NextPage int `json:"nextPage"`
}

func NewSearchResponse(page models.SearchPage) SearchResponse {
	books := make([]BookResponse, 0, len(page.Books))
	for _, book := range page.Books {
		books = append(books, NewBookResponse(book))
	}
	return SearchResponse{Books: books, NextPage: page.NextPage}
}
