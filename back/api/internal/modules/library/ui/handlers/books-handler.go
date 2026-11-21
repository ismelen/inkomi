package handlers

import (
	"errors"
	"net/http"
	"strconv"
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

// HandleSearch serves GET /books/search?q=<query>&by=<title|author>&lang=<language>&fmt=<ext,ext>&page=<n>;
// by defaults to title, fmt to epub and page to 1.
func (b *BooksHandler) HandleSearch(r *http.Request) (*dtos.SearchResponse, error) {
	params := r.URL.Query()
	formats := defaultFormats
	if value := params.Get("fmt"); value != "" {
		formats = strings.Split(value, ",")
	}

	field := models.SearchFieldTitle
	if value := params.Get("by"); value != "" {
		field = models.SearchField(value)
	}

	page := 1
	if value := params.Get("page"); value != "" {
		n, err := strconv.Atoi(value)
		if err != nil {
			return nil, requtil.NewError(http.StatusBadRequest, models.ErrInvalidPage.Error())
		}
		page = n
	}

	result, err := b.searchUC.Execute(r.Context(), params.Get("q"), field, params.Get("lang"), formats, page)
	switch {
	case errors.Is(err, models.ErrEmptyQuery), errors.Is(err, models.ErrInvalidField), errors.Is(err, models.ErrInvalidPage):
		return nil, requtil.NewError(http.StatusBadRequest, err.Error())
	case errors.Is(err, models.ErrNoMirror):
		return nil, requtil.NewError(http.StatusServiceUnavailable, err.Error())
	case err != nil:
		return nil, err
	}

	response := dtos.NewSearchResponse(result)
	return &response, nil
}
