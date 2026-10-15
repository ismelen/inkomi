package handlers

import (
	"errors"
	"net/http"

	"github.com/ismelen/inkomi/back/auth-service/internal/domain/models"
	"github.com/ismelen/inkomi/back/auth-service/internal/infra/api/requtil"
	"github.com/ismelen/inkomi/back/common/httpauth"
)

var errorStatus = []struct {
	err    error
	status int
}{
	{models.ErrInvalidEmail, http.StatusBadRequest},
	{models.ErrWeakPassword, http.StatusBadRequest},
	{models.ErrInvalidDeviceId, http.StatusBadRequest},
	{models.ErrInvalidDropboxFolder, http.StatusBadRequest},
	{models.ErrInvalidFCMToken, http.StatusBadRequest},
	{models.ErrInvalidCredentials, http.StatusUnauthorized},
	{models.ErrInvalidRefreshToken, http.StatusUnauthorized},
	{models.ErrInvalidExternalToken, http.StatusUnauthorized},
	{models.ErrEmailTaken, http.StatusConflict},
	{models.ErrDropboxAccountTaken, http.StatusConflict},
	{models.ErrUserNotFound, http.StatusNotFound},
	{models.ErrDropboxNotLinked, http.StatusNotFound},
	{models.ErrFCMTokenNotFound, http.StatusNotFound},
}

// apiError maps the domain errors to their status. The message is the one of the sentinel,
// so details wrapped around it (e.g. responses of Google or Dropbox) don't reach the client.
func apiError(err error) error {
	for _, e := range errorStatus {
		if errors.Is(err, e.err) {
			return requtil.NewError(e.status, e.err.Error())
		}
	}
	return err
}

func userId(r *http.Request) (int, error) {
	claims, ok := httpauth.GetUserClaims(r.Context())
	if !ok {
		return 0, requtil.NewError(http.StatusUnauthorized, httpauth.ErrInvalidToken.Error())
	}
	return claims.Id, nil
}
