package handlers

import (
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
	"github.com/gorilla/schema"
)

type ConvertHandler struct {
	basePath  string
	convertUC *usecases.ConvertMangaUC
}

func NewConvertHandler(convertUC *usecases.ConvertMangaUC) *ConvertHandler {
	tmp, err := os.MkdirTemp("", "inkomi(*)")
	if err != nil {
		log.Fatal(err)
	}

	return &ConvertHandler{
		basePath:  tmp,
		convertUC: convertUC,
	}
}

var formDecoder = schema.NewDecoder()

func (ch *ConvertHandler) Convert(r *http.Request) (any, error) {
	err := r.ParseMultipartForm(250 << 20)
	if err != nil {
		return nil, err
	}

	req := new(domain.ConvertConfig)
	if err := formDecoder.Decode(req, r.MultipartForm.Value); err != nil {
		return nil, err
	}

	formFiles, err := GetFormFiles(r, "files")
	if err != nil {
		return nil, err
	}

	req.Id = crypto.GetRandomID(6)
	dstPath := filepath.Join(ch.basePath, req.Id)

	chapters, err := converters.FormFilesToChapters(formFiles, filepath.Join(dstPath, "chapters"))
	if err != nil {
		return nil, err
	}
	if len(chapters) == 0 {
		return nil, domain.NewApiError(500, "No files attached")
	}

	for _, c := range chapters {
		log.Println(c.Path)
		for _, p := range c.PagePaths {
			log.Println(p)
		}
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
