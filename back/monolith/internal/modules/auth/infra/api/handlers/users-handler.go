package handlers

import (
	"net/http"

	"github.com/ismelen/inkomi/back/monolith/internal/modules/auth/infra/api/dtos"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/auth/infra/api/requtil"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/auth/usecases/createdropboxfolder"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/auth/usecases/getprofile"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/auth/usecases/linkdropbox"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/auth/usecases/listdropboxfolders"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/auth/usecases/unlinkdropbox"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/auth/usecases/updatedropboxfolder"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/auth/usecases/updatefcmtoken"
)

type UsersHandler struct {
	getProfileUC          *getprofile.GetProfileUC
	updateFCMTokenUC      *updatefcmtoken.UpdateFCMTokenUC
	linkDropboxUC         *linkdropbox.LinkDropboxUC
	updateDropboxFolderUC *updatedropboxfolder.UpdateDropboxFolderUC
	unlinkDropboxUC       *unlinkdropbox.UnlinkDropboxUC
	listDropboxFoldersUC  *listdropboxfolders.ListDropboxFoldersUC
	createDropboxFolderUC *createdropboxfolder.CreateDropboxFolderUC
}

func NewUsersHandler(
	getProfileUC *getprofile.GetProfileUC,
	updateFCMTokenUC *updatefcmtoken.UpdateFCMTokenUC,
	linkDropboxUC *linkdropbox.LinkDropboxUC,
	updateDropboxFolderUC *updatedropboxfolder.UpdateDropboxFolderUC,
	unlinkDropboxUC *unlinkdropbox.UnlinkDropboxUC,
	listDropboxFoldersUC *listdropboxfolders.ListDropboxFoldersUC,
	createDropboxFolderUC *createdropboxfolder.CreateDropboxFolderUC,
) *UsersHandler {
	return &UsersHandler{
		getProfileUC:          getProfileUC,
		updateFCMTokenUC:      updateFCMTokenUC,
		linkDropboxUC:         linkDropboxUC,
		updateDropboxFolderUC: updateDropboxFolderUC,
		unlinkDropboxUC:       unlinkDropboxUC,
		listDropboxFoldersUC:  listDropboxFoldersUC,
		createDropboxFolderUC: createDropboxFolderUC,
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

// HandleListDropboxFolders serves GET /users/me/dropbox/folders?path=, the subfolders of path
// (the root when empty) for the folder picker.
func (u *UsersHandler) HandleListDropboxFolders(r *http.Request) (*dtos.DropboxFoldersResponse, error) {
	id, err := userId(r)
	if err != nil {
		return nil, err
	}

	path, folders, err := u.listDropboxFoldersUC.Execute(r.Context(), id, r.URL.Query().Get("path"))
	if err != nil {
		return nil, apiError(err)
	}
	return dtos.NewDropboxFoldersResponse(path, folders), nil
}

// HandleCreateDropboxFolder serves POST /users/me/dropbox/folders.
func (u *UsersHandler) HandleCreateDropboxFolder(r *http.Request) (*dtos.DropboxFolderEntry, error) {
	id, err := userId(r)
	if err != nil {
		return nil, err
	}
	req, err := requtil.DecodeJSON[dtos.CreateDropboxFolderRequest](r)
	if err != nil {
		return nil, err
	}

	folder, err := u.createDropboxFolderUC.Execute(r.Context(), id, req.Path)
	if err != nil {
		return nil, apiError(err)
	}
	entry := dtos.NewDropboxFolderEntry(*folder)
	return &entry, nil
}
