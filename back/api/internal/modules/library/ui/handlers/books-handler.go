package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/ismelen/inkomi/back/api/internal/modules/library/domain/models"
	"github.com/ismelen/inkomi/back/api/internal/modules/library/ui/dtos"
	"github.com/ismelen/inkomi/back/api/internal/modules/library/ui/requtil"
	"github.com/ismelen/inkomi/back/api/internal/modules/library/usecases/searchbooks"
)

var defaultFormats = []string{"epub"}

type BooksHandler struct {
	searchUC *searchbooks.SearchBooksUC
}

func NewBooksHandler(searchUC *searchbooks.SearchBooksUC) *BooksHandler {
	return &BooksHandler{searchUC: searchUC}
}

// HandleSearch serves GET /books/search?q=<query>&by=<title|author>&lang=<language>&fmt=<ext,ext>;
// by defaults to title and fmt to epub.
func (b *BooksHandler) HandleSearch(r *http.Request) (*[]dtos.BookResponse, error) {
	params := r.URL.Query()
	formats := defaultFormats
	if value := params.Get("fmt"); value != "" {
		formats = strings.Split(value, ",")
	}

	field := models.SearchFieldTitle
	if value := params.Get("by"); value != "" {
		field = models.SearchField(value)
	}

	books, err := b.searchUC.Execute(r.Context(), params.Get("q"), field, params.Get("lang"), formats)
	switch {
	case errors.Is(err, models.ErrEmptyQuery), errors.Is(err, models.ErrInvalidField):
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
