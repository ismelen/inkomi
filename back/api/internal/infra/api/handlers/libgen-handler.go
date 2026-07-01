package handlers

import (
	"ismelen/inkomi/internal/domain/book"
	"ismelen/inkomi/internal/infra/api/requtil"
	"ismelen/inkomi/internal/infra/fs"
	"ismelen/inkomi/internal/shared/uid"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
)

type LibgenHandler struct {
	libgenServ book.LibgenService
}

func NewLibgenHandler(libgenServ book.LibgenService) *LibgenHandler {
	return &LibgenHandler{libgenServ}
}

func (l *LibgenHandler) HandleSearchBook(r *http.Request) (any, error) {
	query := r.URL.Query().Get("q")
	if query == "" {
		return nil, requtil.New(http.StatusBadRequest, "Empty query")
	}

	language := r.URL.Query().Get("lang")

	fmtQuery := r.URL.Query().Get("fmt")
	var formats []string
	if fmtQuery != "" {
		formats = strings.Split(fmtQuery, ",")
	}

	books, err := l.libgenServ.Search(query, language, formats)
	if err != nil {
		return nil, err
	}

	return books, nil
}

func (l *LibgenHandler) HandleDownloadBook(r *http.Request) (any, error) {
	md5 := chi.URLParam(r, "md5")
	result, err := l.libgenServ.Download(md5)
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

	return requtil.FileResponse{
		Path:   path,
		Name:   filepath.Base(path),
		Remove: true,
	}, nil
}
