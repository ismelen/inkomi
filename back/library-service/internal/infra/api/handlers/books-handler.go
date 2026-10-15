package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/ismelen/inkomi/back/library-service/internal/domain/models"
	"github.com/ismelen/inkomi/back/library-service/internal/infra/api/dtos"
	"github.com/ismelen/inkomi/back/library-service/internal/infra/api/requtil"
	"github.com/ismelen/inkomi/back/library-service/internal/usecases/searchbooks"
)

var defaultFormats = []string{"epub"}

type BooksHandler struct {
	searchUC *searchbooks.SearchBooksUC
}

func NewBooksHandler(searchUC *searchbooks.SearchBooksUC) *BooksHandler {
	return &BooksHandler{searchUC: searchUC}
}

// HandleSearch serves GET /books/search?q=<query>&lang=<language>&fmt=<ext,ext>; fmt defaults to epub.
func (b *BooksHandler) HandleSearch(r *http.Request) (*[]dtos.BookResponse, error) {
	params := r.URL.Query()
	formats := defaultFormats
	if value := params.Get("fmt"); value != "" {
		formats = strings.Split(value, ",")
	}

	books, err := b.searchUC.Execute(r.Context(), params.Get("q"), params.Get("lang"), formats)
	switch {
	case errors.Is(err, models.ErrEmptyQuery):
		return nil, requtil.NewError(http.StatusBadRequest, err.Error())
	case errors.Is(err, models.ErrNoMirror):
		return nil, requtil.NewError(http.StatusServiceUnavailable, err.Error())
	case err != nil:
		return nil, err
	}

	response := make([]dtos.BookResponse, 0, len(books))
	for _, book := range books {
		response = append(response, dtos.NewBookResponse(book))
	}
	return &response, nil
}
