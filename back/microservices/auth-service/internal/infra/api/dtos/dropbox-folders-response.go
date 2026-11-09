package dtos

import "github.com/ismelen/inkomi/back/microservices/auth-service/internal/domain/models"

type DropboxFolderEntry struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

// DropboxFoldersResponse lists the subfolders of path, "" being the root.
type DropboxFoldersResponse struct {
	Path    string               `json:"path"`
	Folders []DropboxFolderEntry `json:"folders"`
}

func NewDropboxFolderEntry(folder models.DropboxFolder) DropboxFolderEntry {
	return DropboxFolderEntry{Name: folder.Name, Path: folder.Path}
}

func NewDropboxFoldersResponse(path string, folders []models.DropboxFolder) *DropboxFoldersResponse {
	res := &DropboxFoldersResponse{Path: path, Folders: make([]DropboxFolderEntry, len(folders))}
	for i, folder := range folders {
		res.Folders[i] = NewDropboxFolderEntry(folder)
	}
	return res
}
