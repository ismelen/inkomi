package dtos

import "github.com/ismelen/inkomi/back/library-service/internal/domain/models"

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
		CoverURL:  book.CoverURL,
	}
}
