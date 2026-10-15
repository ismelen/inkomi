package handlers

import (
	"net/http"

	"github.com/ismelen/inkomi/back/auth-service/internal/domain/models"
	"github.com/ismelen/inkomi/back/auth-service/internal/infra/api/dtos"
	"github.com/ismelen/inkomi/back/auth-service/internal/infra/api/requtil"
	"github.com/ismelen/inkomi/back/auth-service/internal/usecases/dropboxlogin"
	"github.com/ismelen/inkomi/back/auth-service/internal/usecases/googlelogin"
	"github.com/ismelen/inkomi/back/auth-service/internal/usecases/login"
	"github.com/ismelen/inkomi/back/auth-service/internal/usecases/logout"
	"github.com/ismelen/inkomi/back/auth-service/internal/usecases/refreshsession"
	"github.com/ismelen/inkomi/back/auth-service/internal/usecases/register"
)

type AuthHandler struct {
	registerUC       *register.RegisterUC
	loginUC          *login.LoginUC
	googleLoginUC    *googlelogin.GoogleLoginUC
	dropboxLoginUC   *dropboxlogin.DropboxLoginUC
	refreshSessionUC *refreshsession.RefreshSessionUC
	logoutUC         *logout.LogoutUC
}

func NewAuthHandler(
	registerUC *register.RegisterUC,
	loginUC *login.LoginUC,
	googleLoginUC *googlelogin.GoogleLoginUC,
	dropboxLoginUC *dropboxlogin.DropboxLoginUC,
	refreshSessionUC *refreshsession.RefreshSessionUC,
	logoutUC *logout.LogoutUC,
) *AuthHandler {
	return &AuthHandler{
		registerUC:       registerUC,
		loginUC:          loginUC,
		googleLoginUC:    googleLoginUC,
		dropboxLoginUC:   dropboxLoginUC,
		refreshSessionUC: refreshSessionUC,
		logoutUC:         logoutUC,
	}
}

// HandleRegister serves POST /auth/register.
func (a *AuthHandler) HandleRegister(r *http.Request) (*dtos.TokensResponse, error) {
	req, err := requtil.DecodeJSON[dtos.RegisterRequest](r)
	if err != nil {
		return nil, err
	}
	return tokens(a.registerUC.Execute(r.Context(), req.Email, req.Password, req.DeviceId))
}

// HandleLogin serves POST /auth/login.
func (a *AuthHandler) HandleLogin(r *http.Request) (*dtos.TokensResponse, error) {
	req, err := requtil.DecodeJSON[dtos.LoginRequest](r)
	if err != nil {
		return nil, err
	}
	return tokens(a.loginUC.Execute(r.Context(), req.Email, req.Password, req.DeviceId))
}

// HandleGoogleLogin serves POST /auth/google.
func (a *AuthHandler) HandleGoogleLogin(r *http.Request) (*dtos.TokensResponse, error) {
	req, err := requtil.DecodeJSON[dtos.GoogleLoginRequest](r)
	if err != nil {
		return nil, err
	}
	return tokens(a.googleLoginUC.Execute(r.Context(), req.IdToken, req.DeviceId))
}

// HandleDropboxLogin serves POST /auth/dropbox.
func (a *AuthHandler) HandleDropboxLogin(r *http.Request) (*dtos.TokensResponse, error) {
	req, err := requtil.DecodeJSON[dtos.DropboxLoginRequest](r)
	if err != nil {
		return nil, err
	}
	return tokens(a.dropboxLoginUC.Execute(r.Context(), req.RefreshToken, req.DeviceId))
}

// HandleRefresh serves POST /auth/refresh.
func (a *AuthHandler) HandleRefresh(r *http.Request) (*dtos.TokensResponse, error) {
	req, err := requtil.DecodeJSON[dtos.RefreshTokenRequest](r)
	if err != nil {
		return nil, err
	}
	return tokens(a.refreshSessionUC.Execute(r.Context(), req.RefreshToken))
}

// HandleLogout serves POST /auth/logout.
func (a *AuthHandler) HandleLogout(r *http.Request) (*struct{}, error) {
	req, err := requtil.DecodeJSON[dtos.RefreshTokenRequest](r)
	if err != nil {
		return nil, err
	}
	return nil, apiError(a.logoutUC.Execute(r.Context(), req.RefreshToken))
}

func tokens(pair *models.TokenPair, err error) (*dtos.TokensResponse, error) {
	if err != nil {
		return nil, apiError(err)
	}
	return dtos.NewTokensResponse(pair), nil
}
