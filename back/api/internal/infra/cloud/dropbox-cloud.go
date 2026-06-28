package cloud

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

type DropboxCloud struct {
	folder string
	token  string
}

func NewDropboxCloud(accessToken, folder string) (*DropboxCloud, error) {
	if folder == "" || accessToken == "" {
		return nil, fmt.Errorf("No folder or access token ")
	}

	return &DropboxCloud{
		folder: folder,
		token:  accessToken,
	}, nil
}

func (d *DropboxCloud) Upload(path string) error {
	filename := filepath.Base(path)

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	dstPath := fmt.Sprintf("%s/%s", d.folder, filename)

	apiArg, err := json.Marshal(map[string]any{
		"path":       dstPath,
		"mode":       "add",
		"autorename": true,
		"mute":       false,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequest(
		http.MethodPost,
		"https://content.dropboxapi.com/2/files/upload",
		file,
	)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+d.token)
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("Dropbox-API-Arg", string(apiArg))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var data map[string]any
	if err := json.Unmarshal(bytes, &data); err != nil {
		return err
	}

	if msg, ok := data["error_summary"]; ok {
		return fmt.Errorf("%s", msg.(string))
	}

	return nil
}
