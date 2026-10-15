package dropbox

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"unicode/utf16"

	"github.com/ismelen/inkomi/back/push-service/internal/domain/models"
	"github.com/ismelen/inkomi/back/push-service/internal/domain/ports"
)

var _ ports.ICloudDrive = (*DropboxCloudDrive)(nil)

const (
	ContentURL = "https://content.dropboxapi.com/2"
	// maxRequestSize is the limit of /files/upload and of each upload session request.
	maxRequestSize = 150 << 20
)

// DropboxCloudDrive uploads files up to chunkSize in a single request and bigger ones
// through an upload session in chunkSize parts.
type DropboxCloudDrive struct {
	client     *http.Client
	contentURL string
	chunkSize  int64
}

func NewDropboxCloudDrive(client *http.Client, contentURL string, chunkSize int64) *DropboxCloudDrive {
	if chunkSize <= 0 || chunkSize > maxRequestSize {
		chunkSize = maxRequestSize
	}
	return &DropboxCloudDrive{
		client:     client,
		contentURL: strings.TrimRight(contentURL, "/"),
		chunkSize:  chunkSize,
	}
}

type commitInfo struct {
	Path       string `json:"path"`
	Mode       string `json:"mode"`
	Autorename bool   `json:"autorename"`
	Mute       bool   `json:"mute"`
}

type cursor struct {
	SessionId string `json:"session_id"`
	Offset    int64  `json:"offset"`
}

type sessionStartArg struct {
	Close bool `json:"close"`
}

type sessionStartResult struct {
	SessionId string `json:"session_id"`
}

type sessionAppendArg struct {
	Cursor cursor `json:"cursor"`
	Close  bool   `json:"close"`
}

type sessionFinishArg struct {
	Cursor cursor     `json:"cursor"`
	Commit commitInfo `json:"commit"`
}

func (d *DropboxCloudDrive) UploadFile(ctx context.Context, account *models.DropboxAccount, path string, filename string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open %s: %w", path, err)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return fmt.Errorf("stat %s: %w", path, err)
	}

	commit := commitInfo{Path: account.Path(filename), Mode: "overwrite"}
	size := info.Size()
	if size <= d.chunkSize {
		return d.call(ctx, account.Token, "/files/upload", commit, io.NewSectionReader(file, 0, size), nil)
	}
	return d.uploadSession(ctx, account.Token, file, size, commit)
}

func (d *DropboxCloudDrive) uploadSession(ctx context.Context, token string, file *os.File, size int64, commit commitInfo) error {
	var started sessionStartResult
	err := d.call(ctx, token, "/files/upload_session/start", sessionStartArg{}, io.NewSectionReader(file, 0, d.chunkSize), &started)
	if err != nil {
		return err
	}

	offset := d.chunkSize
	for size-offset > d.chunkSize {
		arg := sessionAppendArg{Cursor: cursor{SessionId: started.SessionId, Offset: offset}}
		if err := d.call(ctx, token, "/files/upload_session/append_v2", arg, io.NewSectionReader(file, offset, d.chunkSize), nil); err != nil {
			return err
		}
		offset += d.chunkSize
	}

	arg := sessionFinishArg{Cursor: cursor{SessionId: started.SessionId, Offset: offset}, Commit: commit}
	return d.call(ctx, token, "/files/upload_session/finish", arg, io.NewSectionReader(file, offset, size-offset), nil)
}

// call posts body to endpoint with arg in the Dropbox-API-Arg header and decodes the result
// into out when it isn't nil.
func (d *DropboxCloudDrive) call(ctx context.Context, token, endpoint string, arg any, body *io.SectionReader, out any) error {
	apiArg, err := headerSafeJSON(arg)
	if err != nil {
		return fmt.Errorf("encode %s arg: %w", endpoint, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, d.contentURL+endpoint, body)
	if err != nil {
		return fmt.Errorf("create %s request: %w", endpoint, err)
	}
	req.ContentLength = body.Size()
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("Dropbox-API-Arg", apiArg)

	res, err := d.client.Do(req)
	if err != nil {
		return fmt.Errorf("dropbox %s: %w", endpoint, err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return responseError(endpoint, res)
	}
	if out == nil {
		return nil
	}
	if err := json.NewDecoder(res.Body).Decode(out); err != nil {
		return fmt.Errorf("decode dropbox %s: %w", endpoint, err)
	}
	return nil
}

// responseError marks the errors that a retry can't fix (bad request, invalid token, missing
// permissions, conflicts like insufficient space) as models.ErrUploadRejected.
func responseError(endpoint string, res *http.Response) error {
	raw, readErr := io.ReadAll(io.LimitReader(res.Body, 4<<10))
	message := strings.TrimSpace(string(raw))

	var apiErr struct {
		ErrorSummary string `json:"error_summary"`
	}
	if json.Unmarshal(raw, &apiErr) == nil && apiErr.ErrorSummary != "" {
		message = apiErr.ErrorSummary
	}

	err := fmt.Errorf("dropbox %s: status %d: %s", endpoint, res.StatusCode, message)
	switch res.StatusCode {
	case http.StatusBadRequest, http.StatusUnauthorized, http.StatusForbidden, http.StatusConflict:
		err = fmt.Errorf("%w: %w", models.ErrUploadRejected, err)
	}
	return errors.Join(err, readErr)
}

// headerSafeJSON encodes v escaping every non-ASCII character, as Dropbox requires for the
// Dropbox-API-Arg header.
func headerSafeJSON(v any) (string, error) {
	raw, err := json.Marshal(v)
	if err != nil {
		return "", err
	}

	var b strings.Builder
	for _, r := range string(raw) {
		if r < 0x7f {
			b.WriteRune(r)
			continue
		}
		for _, u := range utf16.Encode([]rune{r}) {
			fmt.Fprintf(&b, "\\u%04x", u)
		}
	}
	return b.String(), nil
}
