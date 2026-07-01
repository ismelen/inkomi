package handlers

import (
	"fmt"
	"ismelen/inkomi/internal/domain/convert"
	"ismelen/inkomi/internal/domain/manga"
	"ismelen/inkomi/internal/infra/api/dto"
	"ismelen/inkomi/internal/infra/api/requtil"
	"ismelen/inkomi/internal/infra/api/validation"
	"ismelen/inkomi/internal/infra/fs"
	"ismelen/inkomi/internal/shared/filter"
	"ismelen/inkomi/internal/shared/strutil"
	"ismelen/inkomi/internal/shared/uid"
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
	basePath string
	mangaUC  *usecases.MangaTransactionUC
	epubUC   *usecases.EpubTransactionUC
	decoder  *schema.Decoder
	remoteUC *usecases.RemoteTransactionUC
}

func NewConvertHandler(mangaUC *usecases.MangaTransactionUC, epubUC *usecases.EpubTransactionUC, remoteUC *usecases.RemoteTransactionUC) *TransactionHandler {
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
		basePath: path,
		mangaUC:  mangaUC,
		epubUC:   epubUC,
		decoder:  decoder,
		remoteUC: remoteUC,
	}
}

var filenamesFilter = filter.Use(
	&validation.OnlyOneZipFilter{},
	&validation.SameFormatFilter{},
)

func (ch *TransactionHandler) HandleConvert(r *http.Request) (any, error) {
	modeQuery := r.URL.Query().Get("remote")
	remoteMode := false
	if modeQuery == "true" {
		remoteMode = true
	}

	if err := r.ParseMultipartForm(250 << 20); err != nil {
		return nil, err
	}

	var reqDTO dto.TransactionConfigRequest
	if err := ch.decoder.Decode(&reqDTO, r.MultipartForm.Value); err != nil {
		return nil, err
	}

	if reqDTO.Title != "" {
		if decodedTitle, err := url.QueryUnescape(reqDTO.Title); err == nil {
			reqDTO.Title = decodedTitle
		}
	}

	profile, err := convert.NewProfile(reqDTO.Profile)
	if err != nil {
		return nil, requtil.New(400, err.Error())
	}

	config := &convert.TransactionConfig{
		Author:      reqDTO.Author,
		Title:       reqDTO.Title,
		Profile:     reqDTO.Profile,
		Merge:       reqDTO.Merge,
		Cloud:       reqDTO.Cloud,
		CloudToken:  reqDTO.CloudToken,
		CloudFolder: reqDTO.CloudFolder,
		NotifyToken: reqDTO.NotifyToken,
		ProfileData: profile,
	}

	if remoteMode {
		var md5s []string
		if reqDTO.Md5s == "" {
			return nil, fmt.Errorf("No MD5s specified")
		}
		md5s = strings.Split(reqDTO.Md5s, ",")
		return ch.handleRemoteTransaction(md5s, config)
	}

	children, ext, err := ch.getFilesToProcess(r)
	if err != nil {
		return nil, err
	}

	switch ext {
	case ".zip":
		return nil, requtil.New(400, "Do not send nested zip files")
	case ".epub":
		config.Merge = false
		return ch.handleEpubTransaction(children, config)
	case ".cbz":
		return ch.handleMangaTransaction(children, config)
	default:
		return nil, requtil.New(400, fmt.Sprintf("File format not suported %s", ext))
	}
}

func (ch *TransactionHandler) getFilesToProcess(r *http.Request) ([]string, string, error) {
	files, err := GetFormFiles(r, "files")
	if err != nil {
		return nil, "", err
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
		return nil, "", requtil.New(400, "All files must have same format")
	}

	ctxId := uid.GetRandomID(6)
	tempSavePath := filepath.Join(ch.basePath, "tmp", ctxId)
	defer os.RemoveAll(tempSavePath)

	_, children, err := saveConvertFiles(ext, files, tempSavePath)
	if err != nil {
		return nil, "", err
	}

	if len(children) == 0 {
		return nil, "", requtil.New(400, "No files attached")
	}

	pass, ext = filenamesFilter.Filter(children)
	if !pass {
		return nil, "", requtil.New(400, "All files must have same format")
	}

	return children, ext, nil
}

func (ch *TransactionHandler) handleRemoteTransaction(md5s []string, config *convert.TransactionConfig) (any, error) {
	type TransactionInfo struct {
		md5     string
		config  *convert.TransactionConfig
		dstPath string
	}

	responses := make([]dto.TransactionResponse, 0, len(md5s))
	transactions := make([]TransactionInfo, 0, len(md5s))

	for _, md5 := range md5s {
		id := uid.GetRandomID(6)
		newConfig := config.WithId(id)

		dstPath := filepath.Join(ch.basePath, id)
		if err := os.MkdirAll(dstPath, os.ModePerm); err != nil {
			return nil, err
		}

		transactions = append(transactions, TransactionInfo{
			config:  newConfig,
			md5:     md5,
			dstPath: dstPath,
		})

		responses = append(responses, dto.TransactionResponse{Id: id, Filename: md5})
	}

	go func() {
		for _, tran := range transactions {
			ch.remoteUC.Execute(tran.md5, tran.config, tran.dstPath)
		}
	}()

	return responses, nil
}

func (ch *TransactionHandler) handleEpubTransaction(files []string, config *convert.TransactionConfig) (any, error) {
	type TransactionInfo struct {
		dstPath string
		file    string
		config  *convert.TransactionConfig
	}

	responses := make([]dto.TransactionResponse, 0, len(files))
	transactions := make([]TransactionInfo, 0, len(files))

	for _, file := range files {
		id := uid.GetRandomID(6)
		newConfig := config.WithId(id)

		dstPath := filepath.Join(ch.basePath, id)
		if err := os.MkdirAll(dstPath, os.ModePerm); err != nil {
			return nil, err
		}

		filenameWithExt := filepath.Base(file)
		newConfig.Title = strings.TrimSuffix(filenameWithExt, filepath.Ext(filenameWithExt))

		file, err := fs.CopyFile(file, dstPath)
		if err != nil {
			os.RemoveAll(dstPath)
			return nil, err
		}

		transactions = append(transactions, TransactionInfo{
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

		responses = append(responses, dto.TransactionResponse{Id: id, Title: newConfig.Title, Filename: filename})
	}

	go func() {
		for _, tran := range transactions {
			ch.epubUC.Execute(tran.file, tran.config, tran.dstPath)
		}
	}()

	return responses, nil
}

func (ch *TransactionHandler) handleMangaTransaction(files []string, config *convert.TransactionConfig) (any, error) {
	sort.Slice(files, func(i, j int) bool {
		return strutil.AlphanumericCmp(files[i], files[j])
	})

	type TransactionInfo struct {
		config   *convert.TransactionConfig
		dstPath  string
		chapters []*manga.Chapter
	}

	transactions := make(map[string]*TransactionInfo)
	id := uid.GetRandomID(6)

	for _, file := range files {
		if !config.Merge {
			id = uid.GetRandomID(6)
		}
		dstPath := filepath.Join(ch.basePath, id)
		chapter, err := fs.FileToChapter(file, filepath.Join(dstPath, "chapters"))
		if err != nil {
			return nil, err
		}

		tran, ok := transactions[id]
		if !ok {
			tran = &TransactionInfo{
				config:   config.WithId(id),
				dstPath:  dstPath,
				chapters: []*manga.Chapter{},
			}
		}
		tran.chapters = append(tran.chapters, chapter)
		transactions[id] = tran
	}

	responses := make([]dto.TransactionResponse, 0, len(transactions))
	for id, tran := range transactions {
		tran.config.UpdateTitle(tran.chapters)
		filename := tran.config.Title
		if config.ProfileData.IsKepub {
			filename += ".kepub"
		}
		filename += ".epub"
		responses = append(responses, dto.TransactionResponse{Id: id, Title: tran.config.Title, Filename: filename})
	}

	go func() {
		for _, tran := range transactions {
			defer os.RemoveAll(filepath.Join(tran.dstPath, "chapters"))
			ch.mangaUC.Execute(tran.chapters, tran.config, tran.dstPath)
		}
	}()

	return responses, nil
}

func saveConvertFiles(ext string, files []*multipart.FileHeader, tempSavePath string) (string, []string, error) {
	switch ext {
	case ".zip":
		return fs.UnzipFormZip(files[0], tempSavePath)
	case ".cbz", ".epub":
		return fs.CopyFormFiles(files, tempSavePath)
	default:
		return "", nil, requtil.New(400, fmt.Sprintf("File format not suported %s", ext))
	}
}

func (ch *TransactionHandler) HandleCheckStatus(r *http.Request) (any, error) {
	id := chi.URLParam(r, "id")

	processed, err := ch.mangaUC.CheckProgress(id)
	if err != nil {
		return nil, err
	}

	return map[string]any{"progress": processed}, nil
}

func (ch *TransactionHandler) HandleDownload(r *http.Request) (any, error) {
	id := chi.URLParam(r, "id")

	path, err := ch.mangaUC.GetResultPath(id)
	if err != nil {
		return nil, err
	}

	return requtil.FileResponse{
		Path: path,
		Name: filepath.Base(path),
	}, nil
}

func (ch *TransactionHandler) HandleCancel(r *http.Request) (any, error) {
	id := chi.URLParam(r, "id")
	ch.mangaUC.CancelTransaction(id)
	return map[string]any{}, nil
}
