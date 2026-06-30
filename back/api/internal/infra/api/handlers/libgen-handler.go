package handlers

import (
	"encoding/json"
	"ismelen/inkomi/internal/domain/book"
	"ismelen/inkomi/internal/infra/api/apierr"
	"ismelen/inkomi/internal/infra/api/dto"
	"ismelen/inkomi/internal/infra/fs"
	"ismelen/inkomi/internal/shared/uid"
	"net/http"
	"os"
	"path/filepath"
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
		return nil, apierr.New(http.StatusBadRequest, "Empty query")
	}

	language := r.URL.Query().Get("lang")

	var formats []string
	if fmts := r.URL.Query().Get("fmt"); fmts != "" {
		for _, f := range splitComma(fmts) {
			formats = append(formats, f)
		}
	}

	books, err := l.libgenServ.Search(query, language, formats)
	if err != nil {
		return nil, err
	}

	return books, nil
}

func (l *LibgenHandler) HandleDownloadBook(r *http.Request) (any, error) {
	var reqDTO dto.LibgenDownloadRequest
	if err := json.NewDecoder(r.Body).Decode(&reqDTO); err != nil {
		return nil, apierr.New(http.StatusBadRequest, err.Error())
	}

	if reqDTO.DownloadURL == "" || reqDTO.Title == "" {
		return nil, apierr.New(http.StatusBadRequest, "Download url and title are missing")
	}

	req := book.LibgenDownloadRequest{
		DownloadURL: reqDTO.DownloadURL,
		Title:       reqDTO.Title,
		Extension:   reqDTO.Extension,
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
	id := uid.GetRandomID(6)

	path, err := fs.CopyFromStream(result.Stream, filepath.Join(wd, "books", id, result.Filename))
	if err != nil {
		return nil, err
	}

	return apierr.FileResponse{
		Path:   path,
		Name:   filepath.Base(path),
		Remove: true,
	}, nil
}

func splitComma(s string) []string {
	var out []string
	start := 0
	for i, r := range s {
		if r == ',' {
			out = append(out, s[start:i])
			start = i + 1
		}
	}
	out = append(out, s[start:])
	return out
}
