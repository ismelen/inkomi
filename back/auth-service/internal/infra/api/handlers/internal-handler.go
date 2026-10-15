package handlers

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/ismelen/inkomi/back/auth-service/internal/infra/api/dtos"
	"github.com/ismelen/inkomi/back/auth-service/internal/infra/api/requtil"
	"github.com/ismelen/inkomi/back/auth-service/internal/usecases/getdropboxaccess"
	"github.com/ismelen/inkomi/back/auth-service/internal/usecases/getfcmtoken"
)

// InternalHandler serves the user settings that other services need.
type InternalHandler struct {
	getDropboxAccessUC *getdropboxaccess.GetDropboxAccessUC
	getFCMTokenUC      *getfcmtoken.GetFCMTokenUC
}

func NewInternalHandler(getDropboxAccessUC *getdropboxaccess.GetDropboxAccessUC, getFCMTokenUC *getfcmtoken.GetFCMTokenUC) *InternalHandler {
	return &InternalHandler{getDropboxAccessUC: getDropboxAccessUC, getFCMTokenUC: getFCMTokenUC}
}

// HandleGetDropbox serves GET /internal/users/{userId}/dropbox.
func (i *InternalHandler) HandleGetDropbox(r *http.Request) (*dtos.DropboxAccessResponse, error) {
	id, err := pathUserId(r)
	if err != nil {
		return nil, err
	}

	link, err := i.getDropboxAccessUC.Execute(r.Context(), id)
	if err != nil {
		return nil, apiError(err)
	}
	return dtos.NewDropboxAccessResponse(link), nil
}

// HandleGetFCMToken serves GET /internal/users/{userId}/fcm-token.
func (i *InternalHandler) HandleGetFCMToken(r *http.Request) (*dtos.FCMTokenResponse, error) {
	id, err := pathUserId(r)
	if err != nil {
		return nil, err
	}

	token, err := i.getFCMTokenUC.Execute(r.Context(), id)
	if err != nil {
		return nil, apiError(err)
	}
	return &dtos.FCMTokenResponse{Token: token}, nil
}

func pathUserId(r *http.Request) (int, error) {
	id, err := strconv.Atoi(chi.URLParam(r, "userId"))
	if err != nil || id <= 0 {
		return 0, requtil.NewError(http.StatusBadRequest, "invalid user id")
	}
	return id, nil
}
