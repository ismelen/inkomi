package handlers

import (
	"encoding/json"
	"fmt"
	"ismelen/inkomi/internal/domain"
	"ismelen/inkomi/internal/infra/crypto"
	filesHelper "ismelen/inkomi/internal/infra/files-helper"
	"ismelen/inkomi/internal/infra/libgen"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type LibgenHandler struct {
	libgenServ *libgen.LibgenService
}

func NewLibgenHandler(libgenServ *libgen.LibgenService) *LibgenHandler {
	return &LibgenHandler{libgenServ}
}

func (l *LibgenHandler) HandleSearchBook(r *http.Request) (any, error) {
	query := r.URL.Query().Get("q")
	if query == "" {
		return nil, fmt.Errorf("Empty query")
	}

	language := r.URL.Query().Get("lang")

	formats := []string{}
	if fmts := r.URL.Query().Get("fmt"); fmts != "" {
		formats = append(formats, strings.Split(fmts, ",")...)
	}

	books, err := l.libgenServ.Search(query, language, formats)
	if err != nil {
		return nil, err
	}

	return books, nil
}

func (l *LibgenHandler) HandleDownloadBook(r *http.Request) (any, error) {
	var req domain.LibgenDownloadRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, domain.NewApiError(http.StatusBadRequest, err.Error())
	}

	if req.DownloadURL == "" || req.Title == "" {
		return nil, domain.NewApiError(http.StatusBadRequest, "Download url and title are missing")
	}

	result, err := l.libgenServ.Download(req)
	if err != nil {
		return nil, err
	}
	defer result.Stream.Close()

	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	id := crypto.GetRandomID(6)

	path, err := filesHelper.CopyFromStream(result.Stream, filepath.Join(wd, "books", id))
	if err != nil {
		return nil, err
	}

	return domain.FileResponse{
		Path: path,
		Name: filepath.Base(path),
	}, nil
}
