package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/ismelen/inkomi/back/transaction-manager/internal/infra/api/dtos"
	"github.com/ismelen/inkomi/back/transaction-manager/internal/infra/api/middlewares"
	requtil "github.com/ismelen/inkomi/back/transaction-manager/internal/infra/api/requtils"
	"github.com/ismelen/inkomi/back/transaction-manager/internal/usecases"
)

type UploadsHandler struct {
	newUploadUC usecases.NewUploadRequestUC
	uploadDone  usecases.UploadDoneUC
}

func NewUploadsHandler(
	newUploadUC usecases.NewUploadRequestUC,
	uploadDone usecases.UploadDoneUC,
) *UploadsHandler {
	return &UploadsHandler{
		newUploadUC,
		uploadDone,
	}
}

func (u *UploadsHandler) HandleNewRequest(w http.ResponseWriter, r *http.Request) (*[]dtos.SourceDTO, error) {
	token, _ := middlewares.GetUserClaims(r.Context())

	var data dtos.UploadRequestDTO
	if err := render.DecodeJSON(r.Body, &data); err != nil {
		return nil, requtil.NewError(http.StatusBadRequest, "invalid format")
	}

	sources, err := u.newUploadUC.Execute(r.Context(), data, token.Id)
	return &sources, err
}

func (u *UploadsHandler) HandleOnComplete(w http.ResponseWriter, r *http.Request) (*any, error) {
	token, _ := middlewares.GetUserClaims(r.Context())
	sourceId := chi.URLParam(r, "sourceId")

	err := u.uploadDone.Execute(r.Context(), token.Id, sourceId)
	return nil, err
}
