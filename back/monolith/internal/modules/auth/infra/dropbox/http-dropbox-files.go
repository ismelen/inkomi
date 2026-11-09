package dropbox

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
	"slices"
	"strings"

	"github.com/ismelen/inkomi/back/monolith/internal/modules/auth/domain/models"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/auth/domain/ports"
)

var _ ports.IDropboxFiles = (*HTTPDropboxFiles)(nil)

// HTTPDropboxFiles browses the user's Dropbox with the files endpoints of the API.
type HTTPDropboxFiles struct {
	client *http.Client
	apiURL string
}

func NewHTTPDropboxFiles(client *http.Client, apiURL string) *HTTPDropboxFiles {
	return &HTTPDropboxFiles{client: client, apiURL: strings.TrimRight(apiURL, "/")}
}

type folderEntry struct {
	Tag         string `json:".tag"`
	Name        string `json:"name"`
	PathDisplay string `json:"path_display"`
}

type listFolderResponse struct {
	Entries []folderEntry `json:"entries"`
	Cursor  string        `json:"cursor"`
	HasMore bool          `json:"has_more"`
}

type createFolderResponse struct {
	Metadata folderEntry `json:"metadata"`
}

func (h *HTTPDropboxFiles) ListFolders(ctx context.Context, accessToken, folder string) ([]models.DropboxFolder, error) {
	var page listFolderResponse
	err := h.post(ctx, accessToken, "/2/files/list_folder", map[string]any{
		"path":                           folder,
		"recursive":                      false,
		"include_deleted":                false,
		"include_non_downloadable_files": false,
		"include_mounted_folders":        true,
	}, &page)
	if err != nil {
		return nil, err
	}

	var folders []models.DropboxFolder
	for {
		for _, e := range page.Entries {
			if e.Tag == "folder" {
				folders = append(folders, models.DropboxFolder{Name: e.Name, Path: e.PathDisplay})
			}
		}
		if !page.HasMore {
			break
		}
		cursor := page.Cursor
		page = listFolderResponse{}
		if err := h.post(ctx, accessToken, "/2/files/list_folder/continue", map[string]any{"cursor": cursor}, &page); err != nil {
			return nil, err
		}
	}

	slices.SortFunc(folders, func(a, b models.DropboxFolder) int {
		return strings.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name))
	})
	return folders, nil
}

func (h *HTTPDropboxFiles) CreateFolder(ctx context.Context, accessToken, folder string) (*models.DropboxFolder, error) {
	var res createFolderResponse
	err := h.post(ctx, accessToken, "/2/files/create_folder_v2", map[string]any{"path": folder, "autorename": false}, &res)
	if err != nil {
		return nil, err
	}
	name := res.Metadata.Name
	if name == "" {
		name = path.Base(folder)
	}
	display := res.Metadata.PathDisplay
	if display == "" {
		display = folder
	}
	return &models.DropboxFolder{Name: name, Path: display}, nil
}

// post sends body as JSON and decodes a 200 response into out. A 401 is a revoked or expired
// token; a 409 carries the route error, whose summary tells a missing path from a conflict.
func (h *HTTPDropboxFiles) post(ctx context.Context, accessToken, route string, body, out any) error {
	data, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("encode dropbox %s: %w", route, err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, h.apiURL+route, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("create dropbox %s request: %w", route, err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	res, err := h.client.Do(req)
	if err != nil {
		return fmt.Errorf("dropbox %s: %w", route, err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		// the body only adds detail to the error, so a failed read is ignored
		raw, _ := io.ReadAll(io.LimitReader(res.Body, 1<<10))
		err := fmt.Errorf("dropbox %s: status %d: %s", route, res.StatusCode, strings.TrimSpace(string(raw)))
		switch {
		case res.StatusCode == http.StatusUnauthorized:
			return fmt.Errorf("%w: %w", models.ErrInvalidExternalToken, err)
		case res.StatusCode == http.StatusConflict && bytes.Contains(raw, []byte("not_found")):
			return fmt.Errorf("%w: %w", models.ErrDropboxFolderMissing, err)
		case res.StatusCode == http.StatusConflict && bytes.Contains(raw, []byte("conflict")):
			return fmt.Errorf("%w: %w", models.ErrDropboxFolderExists, err)
		}
		return err
	}

	if err := json.NewDecoder(res.Body).Decode(out); err != nil {
		return fmt.Errorf("decode dropbox %s: %w", route, err)
	}
	return nil
}
