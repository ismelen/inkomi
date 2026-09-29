package handlers

import (
	"context"
	"fmt"
	"io"
	"ismelen/inkomi/internal/domain/convert"
	"ismelen/inkomi/internal/infra/api/dto"
	"ismelen/inkomi/internal/infra/api/requtil"
	"ismelen/inkomi/internal/shared/strutil"
	"ismelen/inkomi/internal/shared/uid"
	"ismelen/inkomi/internal/usecases"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

type TransactionsV2Handler struct {
	tranStore convert.TransactionStoreI
	transPath string
	epubUC    usecases.TransactionUC
	mangaUC   usecases.TransactionUC
	md5UC     usecases.TransactionUC
	notifier  convert.PushNotifier
}

func NewTransactionHandler(
	epubUC usecases.TransactionUC,
	mangaUC usecases.TransactionUC,
	md5UC usecases.TransactionUC,
	tranStore convert.TransactionStoreI,
	notifier convert.PushNotifier,
) *TransactionsV2Handler {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	return &TransactionsV2Handler{
		transPath: filepath.Join(wd, "transactions"),
		epubUC:    epubUC,
		mangaUC:   mangaUC,
		md5UC:     md5UC,
		tranStore: tranStore,
		notifier:  notifier,
	}
}

func (t *TransactionsV2Handler) HandleStartTransaction(r *http.Request) (*dto.TransactionStartResponse, error) {
	var input dto.TransactionStartRequest
	if err := render.DecodeJSON(r.Body, &input); err != nil {
		return nil, requtil.NewError(http.StatusBadRequest, err.Error())
	}

	config, err := input.ToTransactionConfig()
	if err != nil {
		return nil, err
	}

	tran := t.tranStore.StartTransaction(config, t.transPath, func(id string) {
		t.notifier.Send(config.NotifyToken, convert.PushMessage{
			Title:   "Memory allocated",
			Message: "You can now upload files",
			Id:      id,
			Type:    "success",
		})
	})

	return &dto.TransactionStartResponse{
		Id:        tran.Id,
		Timestamp: tran.CreatedAt.Unix(),
		Status:    tran.Status.Get(),
	}, nil
}

func (t *TransactionsV2Handler) getTransaction(r *http.Request) (*convert.Transaction, error) {
	tranId := chi.URLParam(r, "tranId")
	if tranId == "" {
		return nil, requtil.NewError(http.StatusBadRequest, "no transaction id specified")
	}

	return t.tranStore.GetTransaction(tranId)
}

func (t *TransactionsV2Handler) HandleDownload(r *http.Request) (*requtil.FileResponse, error) {
	tran, err := t.getTransaction(r)
	if err != nil {
		return nil, err
	}

	fileId := chi.URLParam(r, "fileId")
	file, err := tran.GetResultFile(fileId)
	if err != nil {
		return nil, err
	}

	return &requtil.FileResponse{
		Path: file.Path,
		Name: file.Name,
	}, nil
}

func (t *TransactionsV2Handler) HandleCancel(r *http.Request) (*any, error) {
	tran, err := t.getTransaction(r)
	if err != nil {
		return nil, err
	}

	tran.Cancel()
	return nil, nil
}

func (t *TransactionsV2Handler) HandleGetStatus(r *http.Request) (*dto.TransactionStatusResponse, error) {
	tran, err := t.getTransaction(r)
	if err != nil {
		return nil, err
	}

	return dto.NewTransactionStatusResponse(tran), nil
}

func (t *TransactionsV2Handler) HandleRetryFile(r *http.Request) (*string, error) {
	fileId := chi.URLParam(r, "fileId")
	tran, err := t.getTransaction(r)
	if err != nil {
		return nil, err
	}

	tran.DeleteItem(fileId)

	return t.HandleAttachFile(r)
}

func (t *TransactionsV2Handler) HandleAttachFile(r *http.Request) (*string, error) {
	tran, err := t.getTransaction(r)
	if err != nil {
		return nil, err
	}

	if tran.Config.Type == "md5" {
		return t.handleAttachMd5(tran, r)
	}

	file, header, err := requtil.GetFormFile(r)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	if tran.Config.Type == "epub" {
		if ext := filepath.Ext(header.Filename); ext != ".epub" {
			return nil, requtil.NewError(
				http.StatusBadRequest,
				fmt.Sprintf("%s file type not suported, only %s", ext, tran.Config.Type),
			)
		}
	}

	tranFileId := uid.GetRandomID(6)
	decodedFilename, err := url.QueryUnescape(header.Filename)
	if err != nil {
		return nil, err
	}

	filename := strutil.NormalizeString(decodedFilename)
	dirPath := filepath.Join(t.transPath, tran.Id, tranFileId)
	dstPath := filepath.Join(dirPath, filename)

	if err := os.MkdirAll(dirPath, os.ModeAppend); err != nil {
		return nil, err
	}

	dst, err := os.Create(dstPath)
	if err != nil {
		return nil, err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return nil, err
	}

	tranFile := convert.NewTransactionFile(tranFileId, filename, dstPath, header.Size)
	fileId, err := tran.AttachFile(tranFile)

	ctx := context.WithoutCancel(r.Context())
	go func() {
		switch tran.Config.Type {
		case "cbz":
			t.mangaUC.Execute(ctx, tranFile, tran, t.transPath)
		case "epub":
			t.epubUC.Execute(ctx, tranFile, tran, t.transPath)
		}
	}()

	return &fileId, err
}

func (t *TransactionsV2Handler) handleAttachMd5(tran *convert.Transaction, r *http.Request) (*string, error) {
	var input dto.Md5TransactionAttachRequest
	if err := render.DecodeJSON(r.Body, &input); err != nil {
		return nil, err
	}

	tranFile := convert.NewTransactionFile(uid.GetRandomID(6), input.Title, input.Md5, 0)
	fileId, err := tran.AttachFile(tranFile)

	ctx := context.WithoutCancel(r.Context())
	go func() { t.md5UC.Execute(ctx, tranFile, tran, t.transPath) }()
	return &fileId, err
}
