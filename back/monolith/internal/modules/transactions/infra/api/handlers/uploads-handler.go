package handlers

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/transactions/domain/models"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/transactions/infra/api/dtos"
	requtil "github.com/ismelen/inkomi/back/monolith/internal/modules/transactions/infra/api/requtils"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/transactions/usecases/getdownload"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/transactions/usecases/listuploads"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/transactions/usecases/newupload"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/transactions/usecases/retryupload"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/transactions/usecases/uploaddone"
	"github.com/ismelen/inkomi/back/monolith/internal/platform/httpauth"
)

type UploadsHandler struct {
	newUploadUC   *newupload.NewUploadRequestUC
	uploadDoneUC  *uploaddone.UploadDoneUC
	listUploadsUC *listuploads.ListUploadsUC
	getDownloadUC *getdownload.GetDownloadUC
	retryUploadUC *retryupload.RetryUploadUC
}

func NewUploadsHandler(
	newUploadUC *newupload.NewUploadRequestUC,
	uploadDoneUC *uploaddone.UploadDoneUC,
	listUploadsUC *listuploads.ListUploadsUC,
	getDownloadUC *getdownload.GetDownloadUC,
	retryUploadUC *retryupload.RetryUploadUC,
) *UploadsHandler {
	return &UploadsHandler{
		newUploadUC:   newUploadUC,
		uploadDoneUC:  uploadDoneUC,
		listUploadsUC: listUploadsUC,
		getDownloadUC: getDownloadUC,
		retryUploadUC: retryUploadUC,
	}
}

func (u *UploadsHandler) HandleNewRequest(w http.ResponseWriter, r *http.Request) (*[]dtos.SourceDTO, error) {
	token, _ := httpauth.GetUserClaims(r.Context())

	var data dtos.UploadRequestDTO
	if err := render.DecodeJSON(r.Body, &data); err != nil {
		return nil, requtil.NewError(http.StatusBadRequest, "invalid format")
	}

	sources, err := u.newUploadUC.Execute(r.Context(), data, token.Id)
	if errors.Is(err, models.ErrInsufficientStorage) {
		return nil, requtil.NewError(http.StatusInsufficientStorage, err.Error())
	}
	if err != nil {
		return nil, err
	}
	for _, source := range sources {
		log.Printf("upload request user %d: %s %s %q with %d items", token.Id, source.Type, source.Id, source.Title, len(source.Items))
	}
	return &sources, nil
}

func (u *UploadsHandler) HandleOnComplete(w http.ResponseWriter, r *http.Request) (*any, error) {
	token, _ := httpauth.GetUserClaims(r.Context())
	sourceId := chi.URLParam(r, "sourceId")

	if err := u.uploadDoneUC.Execute(r.Context(), token.Id, sourceId); err != nil {
		return nil, err
	}
	log.Printf("upload done user %d: %s", token.Id, sourceId)
	return nil, nil
}

func (u *UploadsHandler) HandleList(w http.ResponseWriter, r *http.Request) (*dtos.UploadListResponse, error) {
	token, _ := httpauth.GetUserClaims(r.Context())
	query := r.URL.Query()

	limit := models.DefaultSourceListLimit
	if raw := query.Get("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			return nil, requtil.NewError(http.StatusBadRequest, models.ErrInvalidLimit.Error())
		}
		limit = parsed
	}

	var updatedSince *time.Time
	if raw := query.Get("updatedSince"); raw != "" {
		parsed, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			return nil, requtil.NewError(http.StatusBadRequest, "invalid updatedSince")
		}
		updatedSince = &parsed
	}

	filter, err := models.NewSourceListFilter(limit, query.Get("cursor"), updatedSince)
	if err != nil {
		return nil, requtil.NewError(http.StatusBadRequest, err.Error())
	}

	page, err := u.listUploadsUC.Execute(r.Context(), token.Id, *filter)
	if err != nil {
		return nil, err
	}

	return dtos.NewUploadListResponse(page), nil
}

func (u *UploadsHandler) HandleDownload(w http.ResponseWriter, r *http.Request) (*dtos.DownloadResponse, error) {
	token, _ := httpauth.GetUserClaims(r.Context())
	sourceId := chi.URLParam(r, "sourceId")

	download, err := u.getDownloadUC.Execute(r.Context(), token.Id, sourceId)
	switch {
	case errors.Is(err, models.ErrSourceNotFound), errors.Is(err, models.ErrFileNotAvailable):
		return nil, requtil.NewError(http.StatusNotFound, err.Error())
	case errors.Is(err, models.ErrSourceNotReady):
		return nil, requtil.NewError(http.StatusConflict, err.Error())
	case err != nil:
		return nil, err
	}

	return dtos.NewDownloadResponse(download), nil
}

// HandleRetry serves POST /uploads/{sourceId}/retry: a new upload url for a file whose upload
// failed, keeping it in its folder.
func (u *UploadsHandler) HandleRetry(w http.ResponseWriter, r *http.Request) (*dtos.RetryUploadResponse, error) {
	token, _ := httpauth.GetUserClaims(r.Context())
	sourceId := chi.URLParam(r, "sourceId")

	url, err := u.retryUploadUC.Execute(r.Context(), token.Id, sourceId)
	switch {
	case errors.Is(err, models.ErrSourceNotFound):
		return nil, requtil.NewError(http.StatusNotFound, err.Error())
	case errors.Is(err, models.ErrSourceNotRetryable):
		return nil, requtil.NewError(http.StatusConflict, err.Error())
	case err != nil:
		return nil, err
	}
	log.Printf("upload retry user %d: %s", token.Id, sourceId)
	return &dtos.RetryUploadResponse{Id: sourceId, Url: url}, nil
}
