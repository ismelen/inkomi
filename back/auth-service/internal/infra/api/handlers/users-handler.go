package handlers

import (
	"net/http"

	"github.com/ismelen/inkomi/back/auth-service/internal/infra/api/dtos"
	"github.com/ismelen/inkomi/back/auth-service/internal/infra/api/requtil"
	"github.com/ismelen/inkomi/back/auth-service/internal/usecases/getprofile"
	"github.com/ismelen/inkomi/back/auth-service/internal/usecases/linkdropbox"
	"github.com/ismelen/inkomi/back/auth-service/internal/usecases/unlinkdropbox"
	"github.com/ismelen/inkomi/back/auth-service/internal/usecases/updatedropboxfolder"
	"github.com/ismelen/inkomi/back/auth-service/internal/usecases/updatefcmtoken"
)

type UsersHandler struct {
	getProfileUC          *getprofile.GetProfileUC
	updateFCMTokenUC      *updatefcmtoken.UpdateFCMTokenUC
	linkDropboxUC         *linkdropbox.LinkDropboxUC
	updateDropboxFolderUC *updatedropboxfolder.UpdateDropboxFolderUC
	unlinkDropboxUC       *unlinkdropbox.UnlinkDropboxUC
}

func NewUsersHandler(
	getProfileUC *getprofile.GetProfileUC,
	updateFCMTokenUC *updatefcmtoken.UpdateFCMTokenUC,
	linkDropboxUC *linkdropbox.LinkDropboxUC,
	updateDropboxFolderUC *updatedropboxfolder.UpdateDropboxFolderUC,
	unlinkDropboxUC *unlinkdropbox.UnlinkDropboxUC,
) *UsersHandler {
	return &UsersHandler{
		getProfileUC:          getProfileUC,
		updateFCMTokenUC:      updateFCMTokenUC,
		linkDropboxUC:         linkDropboxUC,
		updateDropboxFolderUC: updateDropboxFolderUC,
		unlinkDropboxUC:       unlinkDropboxUC,
	}
}

// HandleGetProfile serves GET /users/me.
func (u *UsersHandler) HandleGetProfile(r *http.Request) (*dtos.ProfileResponse, error) {
	id, err := userId(r)
	if err != nil {
		return nil, err
	}

	profile, err := u.getProfileUC.Execute(r.Context(), id)
	if err != nil {
		return nil, apiError(err)
	}
	return dtos.NewProfileResponse(profile), nil
}

// HandleUpdateFCMToken serves PUT /users/me/fcm-token; an empty token removes it.
func (u *UsersHandler) HandleUpdateFCMToken(r *http.Request) (*struct{}, error) {
	id, err := userId(r)
	if err != nil {
		return nil, err
	}
	req, err := requtil.DecodeJSON[dtos.FCMTokenRequest](r)
	if err != nil {
		return nil, err
	}
	return nil, apiError(u.updateFCMTokenUC.Execute(r.Context(), id, req.Token))
}

// HandleLinkDropbox serves PUT /users/me/dropbox with the Dropbox refresh token of the client.
func (u *UsersHandler) HandleLinkDropbox(r *http.Request) (*struct{}, error) {
	id, err := userId(r)
	if err != nil {
		return nil, err
	}
	req, err := requtil.DecodeJSON[dtos.DropboxLinkRequest](r)
	if err != nil {
		return nil, err
	}
	return nil, apiError(u.linkDropboxUC.Execute(r.Context(), id, req.RefreshToken))
}

// HandleUpdateDropboxFolder serves PUT /users/me/dropbox/folder.
func (u *UsersHandler) HandleUpdateDropboxFolder(r *http.Request) (*dtos.DropboxFolderResponse, error) {
	id, err := userId(r)
	if err != nil {
		return nil, err
	}
	req, err := requtil.DecodeJSON[dtos.DropboxFolderRequest](r)
	if err != nil {
		return nil, err
	}

	folder, err := u.updateDropboxFolderUC.Execute(r.Context(), id, req.Folder)
	if err != nil {
		return nil, apiError(err)
	}
	return &dtos.DropboxFolderResponse{Folder: folder}, nil
}

// HandleUnlinkDropbox serves DELETE /users/me/dropbox.
func (u *UsersHandler) HandleUnlinkDropbox(r *http.Request) (*struct{}, error) {
	id, err := userId(r)
	if err != nil {
		return nil, err
	}
	return nil, apiError(u.unlinkDropboxUC.Execute(r.Context(), id))
}
