package dtos

import (
	"net/url"

	"github.com/ismelen/inkomi/back/api/internal/modules/library/domain/models"
)

type BookResponse struct {
	Title     string `json:"title"`
	Author    string `json:"author"`
	Pages     int    `json:"pages"`
	Language  string `json:"language"`
	Extension string `json:"extension"`
	MD5       string `json:"md5"`
	CoverURL  string `json:"coverUrl"`
}

func NewBookResponse(book models.Book) BookResponse {
	return BookResponse{
		Title:     book.Title,
		Author:    book.Author,
		Pages:     book.Pages,
		Language:  book.Language,
		Extension: book.Extension,
		MD5:       book.MD5,
		CoverURL:  coverURL(book.CoverURL),
	}
}

// coverURL points the cover to the api's /books/cover, which fetches it from the mirror: the
// clients may not reach the libgen hosts themselves.
func coverURL(mirrorURL string) string {
	if mirrorURL == "" {
		return ""
	}
	return "/books/cover?url=" + url.QueryEscape(mirrorURL)
}
