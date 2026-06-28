package handlers

import (
	"fmt"
	"ismelen/inkomi/internal/domain"
	"ismelen/inkomi/internal/infra/converters"
	"ismelen/inkomi/internal/infra/crypto"
	filesHelper "ismelen/inkomi/internal/infra/files-helper"
	filesFilter "ismelen/inkomi/internal/infra/filters/files-filter"
	"ismelen/inkomi/internal/infra/helpers"
	"ismelen/inkomi/internal/infra/state"
	"ismelen/inkomi/internal/ports"
	"ismelen/inkomi/internal/usecases"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/schema"
)

type TransactionHandler struct {
	basePath     string
	pushNotifier ports.PushNotifier
	decoder      *schema.Decoder
}

func NewConvertHandler(pushNotifier ports.PushNotifier) *TransactionHandler {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	path := filepath.Join(wd, "transactions")
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		log.Fatal(err)
	}

	decoder := schema.NewDecoder()
	decoder.IgnoreUnknownKeys(true)
	decoder.SetAliasTag("form")

	return &TransactionHandler{
		basePath:     path,
		pushNotifier: pushNotifier,
		decoder:      decoder,
	}
}

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
		if decodedName, err := url.QueryUnescape(file.Filename); err == nil {
			file.Filename = decodedName
		}
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
	if err := ch.decoder.Decode(config, r.MultipartForm.Value); err != nil {
		return nil, err
	}

	if config.Title != "" {
		if decodedTitle, err := url.QueryUnescape(config.Title); err == nil {
			config.Title = decodedTitle
		}
	}

	profile, err := domain.NewProfile(config.Profile)
	if err != nil {
		return nil, err
	}
	config.ProfileData = profile

	switch ext {
	case ".zip":
		return nil, domain.NewApiError(400, "Do not send nested zip files")
	case ".epub":
		config.Merge = false
		return ch.handleEpubTransaction(children, config)
	case ".cbz":
		return ch.handleMangaTransaction(children, config)
	default:
		return nil, domain.NewApiError(400, fmt.Sprintf("File format not suported %s", ext))
	}
}

func (ch *TransactionHandler) handleEpubTransaction(files []string, config *domain.TransactionConfig) (any, error) {
	type TransactinoInfo struct {
		dstPath string
		file    string
		config  *domain.TransactionConfig
	}

	responses := make([]domain.TransactionResponseDTO, 0, len(files))
	transactions := make([]TransactinoInfo, 0, len(files))
	for _, file := range files {
		id := crypto.GetRandomID(6)
		newConfig := config.WithId(id)

		dstPath := filepath.Join(ch.basePath, id)
		if err := os.MkdirAll(dstPath, os.ModePerm); err != nil {
			return nil, err
		}

		filenameWithExt := filepath.Base(file)
		newConfig.Title = strings.TrimSuffix(filenameWithExt, filepath.Ext(filenameWithExt))

		file, err := filesHelper.CopyFile(file, dstPath)
		if err != nil {
			os.RemoveAll(dstPath)
			return nil, err
		}

		transactions = append(transactions, TransactinoInfo{
			dstPath: dstPath,
			config:  newConfig,
			file:    file,
		})

		filename := filepath.Base(file)

		if config.ProfileData.IsKepub {
			ext := filepath.Ext(file)
			name := strings.TrimSuffix(filename, ext)
			filename = name + ".kepub" + ext
		}

		responses = append(responses, domain.NewTransactionResponseDTO(id, newConfig.Title, filename))
	}

	transactionUC := usecases.NewEpubTransactionUC(ch.pushNotifier)
	go func() {
		for _, tran := range transactions {
			transactionUC.Execute(tran.file, tran.config, tran.dstPath)
		}
	}()

	return responses, nil
}

func (ch *TransactionHandler) handleMangaTransaction(files []string, config *domain.TransactionConfig) (any, error) {
	sort.Slice(files, func(i, j int) bool {
		return helpers.AlphanumericCmp(files[i], files[j])
	})

	type TransactinoInfo struct {
		config   *domain.TransactionConfig
		dstPath  string
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
			tran = &TransactinoInfo{
				config:   config.WithId(id),
				dstPath:  dstPath,
				chapters: []*domain.Chapter{},
			}
		}
		tran.chapters = append(tran.chapters, chapter)
		transactions[id] = tran
	}

	responses := make([]domain.TransactionResponseDTO, 0, len(transactions))
	for id, tran := range transactions {
		tran.config.UpdateTitle(tran.chapters)
		filename := tran.config.Title
		if config.ProfileData.IsKepub {
			filename += ".kepub"
		}
		filename += ".epub"

		responses = append(responses, domain.NewTransactionResponseDTO(id, tran.config.Title, filename))
	}

	transactionUC := usecases.NewMangaTransactionUC(ch.pushNotifier)
	go func() {
		for _, tran := range transactions {
			defer os.RemoveAll(filepath.Join(tran.dstPath, "chapters"))
			transactionUC.Execute(tran.chapters, tran.config, tran.dstPath)
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
