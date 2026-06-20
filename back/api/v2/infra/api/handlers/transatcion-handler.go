package handlers

import (
	"fmt"
	"ismelen/ermc/v2/domain"
	"ismelen/ermc/v2/infra/converters"
	"ismelen/ermc/v2/infra/crypto"
	filesHelper "ismelen/ermc/v2/infra/files-helper"
	filesFilter "ismelen/ermc/v2/infra/filters/files-filter"
	"ismelen/ermc/v2/infra/helpers"
	"ismelen/ermc/v2/infra/state"
	"ismelen/ermc/v2/usecases"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"sort"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/schema"
)

type TransactionHandler struct {
	basePath  string
	convertUC *usecases.ConvertMangaUC
}

func NewConvertHandler(convertUC *usecases.ConvertMangaUC) *TransactionHandler {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	path := filepath.Join(wd, "transactions")
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		log.Fatal(err)
	}

	return &TransactionHandler{
		basePath:  path,
		convertUC: convertUC,
	}
}

var formDecoder = schema.NewDecoder()
var filenamesFilter = filesFilter.Use(
	&filesFilter.OnlyOneZipFilter{},
	&filesFilter.SameFormatFilter{},
)

func (ch *TransactionHandler) HandleConvert(r *http.Request) (any, error) {
	if err := r.ParseMultipartForm(250 << 20); err != nil {
		return nil, err
	}

	files, err := GetFormFiles(r, "files")
	if err != nil {
		return nil, err
	}

	filenames := make([]string, 0, len(files))
	for _, file := range files {
		filenames = append(filenames, file.Filename)
	}

	pass, ext := filenamesFilter.Filter(filenames)
	if !pass {
		return nil, domain.NewApiError(400, "All files must have same format")
	}

	ctxId := crypto.GetRandomID(6)
	tempSavePath := filepath.Join(ch.basePath, "tmp", ctxId)
	defer os.RemoveAll(tempSavePath)

	_, children, err := saveConvertFiles(ext, files, tempSavePath)
	if err != nil {
		return nil, err
	}

	if len(children) == 0 {
		return nil, domain.NewApiError(400, "No files attached")
	}

	pass, ext = filenamesFilter.Filter(children)
	if !pass {
		return nil, domain.NewApiError(400, "All files must have same format")
	}

	config := new(domain.TransactionConfig)
	if err := formDecoder.Decode(config, r.MultipartForm.Value); err != nil {
		return nil, err
	}

	switch ext {
	case ".zip":
		return nil, domain.NewApiError(400, "Do not send nested zip files")
	case ".epub":
	case ".cbz":
		log.Println("cbz")
		return ch.handleConvertManga(children, config)
	default:
		return nil, domain.NewApiError(400, fmt.Sprintf("File format not suported %s", ext))
	}

	// req.Id = crypto.GetRandomID(6)
	// dstPath := filepath.Join(ch.basePath, req.Id)

	// chapters, err := converters.FormFilesToChapters(formFiles, filepath.Join(dstPath, "chapters"))
	// if err != nil {
	// 	return nil, err
	// }
	// if len(chapters) == 0 {
	// 	return nil, domain.NewApiError(500, "No files attached")
	// }

	// //! TODO: change implementation to get better titles
	// if req.Title == "" {
	// 	req.Title = filepath.Base(chapters[0].Path)
	// }

	// go ch.convertUC.Execute(chapters, req, dstPath)
	// return domain.NewConvertResponseDTO(req.Id, req.Title), nil

	return "", nil
}

func (ch *TransactionHandler) handleConvertManga(files []string, config *domain.TransactionConfig) (any, error) {
	sort.Slice(files, func(i, j int) bool {
		return helpers.AlphanumericCmp(files[i], files[j])
	})

	type TransactinoInfo struct {
		config   *domain.TransactionConfig
		filepath string
		chapters []*domain.Chapter
	}

	transactions := make(map[string]*TransactinoInfo)
	id := crypto.GetRandomID(6)

	for _, file := range files {
		if !config.Merge {
			id = crypto.GetRandomID(6)
		}
		dstPath := filepath.Join(ch.basePath, id)
		chapter, err := converters.FileToChapter(file, filepath.Join(dstPath, "chapters"))
		if err != nil {
			return nil, err
		}

		tran, ok := transactions[id]
		if !ok {
			log.Println("new transaction")
			tran = &TransactinoInfo{
				config:   config.WithId(id),
				filepath: dstPath,
				chapters: []*domain.Chapter{},
			}
		}
		tran.chapters = append(tran.chapters, chapter)
		transactions[id] = tran
	}

	responses := make([]domain.TransactionResponseDTO, 0, len(transactions))
	for id, tran := range transactions {
		tran.config.UpdateTitle(tran.chapters)
		responses = append(responses, domain.NewTransactionResponseDTO(id, tran.config.Title))
	}

	go func() {
		for _, tran := range transactions {
			defer os.RemoveAll(filepath.Join(tran.filepath, "chapters"))
			ch.convertUC.Execute(tran.chapters, tran.config, tran.filepath)
		}
	}()

	return responses, nil
}

func saveConvertFiles(ext string, files []*multipart.FileHeader, tempSavePath string) (string, []string, error) {
	switch ext {
	case ".zip":
		return filesHelper.UnzipFormZip(files[0], tempSavePath)
	case ".cbz", ".epub":
		return filesHelper.CopyFormFiles(files, tempSavePath)
	default:
		return "", nil, domain.NewApiError(400, fmt.Sprintf("File format not suported %s", ext))
	}
}

func handleKepubify(files []*multipart.FileHeader) (any, error) {
	log.Println(files)
	return map[string]any{}, nil
}

func (ch *TransactionHandler) HandleCheckStatus(r *http.Request) (any, error) {
	id := chi.URLParam(r, "id")
	stateMng := state.GetManager()

	processed, err := stateMng.CheckProgress(id)
	if err != nil {
		return nil, err
	}

	return map[string]any{"progress": processed}, nil
}

func (ch *TransactionHandler) HandleDownload(r *http.Request) (any, error) {
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

func (ch *TransactionHandler) HandleCancel(r *http.Request) (any, error) {
	id := chi.URLParam(r, "id")

	stateMng := state.GetManager()
	stateMng.Cancel(id)

	return map[string]any{}, nil
}
