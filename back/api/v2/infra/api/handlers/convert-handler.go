package handlers

import (
	"encoding/json"
	"ismelen/ermc/v2/domain"
	"ismelen/ermc/v2/infra/converters"
	"ismelen/ermc/v2/infra/crypto"
	"ismelen/ermc/v2/infra/state"
	"ismelen/ermc/v2/usecases"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-chi/chi/v5"
)

type ConvertHandler struct {
	basePath  string
	convertUC *usecases.ConvertMangaUC
}

func NewConvertHandler(convertUC *usecases.ConvertMangaUC) *ConvertHandler {
	tmp, err := os.MkdirTemp("", "ERMC(*)")
	if err != nil {
		log.Fatal(err)
	}

	return &ConvertHandler{
		basePath:  tmp,
		convertUC: convertUC,
	}
}

func (ch *ConvertHandler) Convert(r *http.Request) (any, error) {
	req := new(domain.ConvertConfig)
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, err
	}

	formFiles, err := GetFormFiles(r, "files")
	if err != nil {
		return nil, err
	}

	req.Id = crypto.GetRandomID(6)
	dstPath := filepath.Join(ch.basePath, req.Id)

	chapters, err := converters.FilesToChapters(formFiles, filepath.Join(dstPath, "chapters"))
	if err != nil {
		return nil, err
	}

	if req.Title == "" {
		req.Title = filepath.Base(chapters[0].Path)
	}

	go ch.convertUC.Execute(chapters, req, dstPath)
	return map[string]any{
		"id":    req.Id,
		"title": req.Title,
	}, nil
}

func (ch *ConvertHandler) CheckStatus(r *http.Request) (any, error) {
	id := chi.URLParam(r, "id")
	stateMng := state.GetManager()

	processed, err := stateMng.CheckProgress(id)
	if err != nil {
		return nil, err
	}

	return map[string]any{"progress": processed}, nil
}

func (ch *ConvertHandler) Download(r *http.Request) (any, error) {
	id := chi.URLParam(r, "id")
	stateMng := state.GetManager()

	path, err := stateMng.GetResultPath(id)
	if err != nil {
		return nil, err
	}

	return domain.FileResponse{
		Path: path,
		Name: filepath.Base(path),
	}, nil
}

func (ch *ConvertHandler) Cancel(r *http.Request) (any, error) {
	id := chi.URLParam(r, "id")

	stateMng := state.GetManager()
	stateMng.Cancel(id)

	return map[string]any{}, nil
}
