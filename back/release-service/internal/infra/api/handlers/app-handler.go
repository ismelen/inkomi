package handlers

import (
	"errors"
	"net/http"

	"github.com/ismelen/inkomi/back/release-service/internal/domain/models"
	"github.com/ismelen/inkomi/back/release-service/internal/infra/api/dtos"
	"github.com/ismelen/inkomi/back/release-service/internal/infra/api/requtil"
	"github.com/ismelen/inkomi/back/release-service/internal/usecases/getlatestrelease"
)

const DownloadPath = "/app/download"

type AppHandler struct {
	latestUC *getlatestrelease.GetLatestReleaseUC
}

func NewAppHandler(latestUC *getlatestrelease.GetLatestReleaseUC) *AppHandler {
	return &AppHandler{latestUC: latestUC}
}

// HandleLatest serves GET /app/latest: the version the app compares with its own.
func (a *AppHandler) HandleLatest(r *http.Request) (*dtos.ReleaseResponse, error) {
	release, err := a.latestUC.Execute(r.Context())
	if err != nil {
		return nil, apiError(err)
	}
	return dtos.NewReleaseResponse(release, DownloadPath), nil
}

// HandleDownload serves GET /app/download: a redirect to the apk of the latest release.
func (a *AppHandler) HandleDownload(r *http.Request) (*requtil.Redirect, error) {
	release, err := a.latestUC.Execute(r.Context())
	if err != nil {
		return nil, apiError(err)
	}
	return &requtil.Redirect{URL: release.APK.URL}, nil
}

func apiError(err error) error {
	switch {
	case errors.Is(err, models.ErrReleaseNotFound), errors.Is(err, models.ErrAPKNotFound):
		return requtil.NewError(http.StatusNotFound, models.ErrReleaseNotFound.Error())
	case errors.Is(err, models.ErrSourceUnavailable):
		return requtil.NewError(http.StatusServiceUnavailable, models.ErrSourceUnavailable.Error())
	default:
		return err
	}
}
