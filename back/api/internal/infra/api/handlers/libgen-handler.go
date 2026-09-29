package handlers

import (
	"ismelen/inkomi/internal/domain/book"
	"ismelen/inkomi/internal/infra/api/requtil"
	"ismelen/inkomi/internal/infra/fs"
	"ismelen/inkomi/internal/shared/uid"
	"ismelen/inkomi/internal/usecases"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
)

type LibgenHandler struct {
	checkUC    *usecases.CheckBooksSourceUC
	searchUC   *usecases.SearchBookUC
	downloadUC *usecases.DownloadBookUC
}

func NewLibgenHandler(
	checkUC *usecases.CheckBooksSourceUC,
	searchUC *usecases.SearchBookUC,
	downloadUC *usecases.DownloadBookUC,
) *LibgenHandler {
	return &LibgenHandler{
		checkUC:    checkUC,
		searchUC:   searchUC,
		downloadUC: downloadUC,
	}
}

type CheckMirrorResponse struct {
	Active bool `json:"active"`
}

func (l *LibgenHandler) HandleCheckMirror(r *http.Request) (*CheckMirrorResponse, error) {
	active, err := l.checkUC.Execute(r.Context())
	if err != nil {
		return nil, err
	}

	return &CheckMirrorResponse{Active: active}, nil
}

func (l *LibgenHandler) HandleSearchBook(r *http.Request) (*[]book.Book, error) {
	query := r.URL.Query().Get("q")
	if query == "" {
		return nil, requtil.NewError(http.StatusBadRequest, "Empty query")
	}

	language := r.URL.Query().Get("lang")

	fmtQuery := r.URL.Query().Get("fmt")
	formats := []string{"epub"}
	if fmtQuery != "" {
		formats = strings.Split(fmtQuery, ",")
	}

	books, err := l.searchUC.Execute(r.Context(), query, language, formats)
	if err != nil {
		return nil, err
	}

	return &books, nil
}

func (l *LibgenHandler) HandleDownloadBook(r *http.Request) (*requtil.FileResponse, error) {
	md5 := chi.URLParam(r, "md5")
	result, err := l.downloadUC.Execute(r.Context(), md5, 3)
	if err != nil {
		return nil, err
	}
	defer result.Stream.Close()

	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	id := uid.GetRandomID(6)

	path, err := fs.CopyFromStream(result.Stream, filepath.Join(wd, "books", id, result.Filename))
	if err != nil {
		return nil, err
	}

	return &requtil.FileResponse{
		Path:   path,
		Name:   filepath.Base(path),
		Remove: true,
	}, nil
}
