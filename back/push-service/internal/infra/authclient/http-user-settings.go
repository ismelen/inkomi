package authclient

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/ismelen/inkomi/back/push-service/internal/domain/models"
	"github.com/ismelen/inkomi/back/push-service/internal/domain/ports"
)

var _ ports.IUserSettings = (*HTTPUserSettings)(nil)

type ITokenSigner interface {
	Sign(audience string) (string, error)
}

// HTTPUserSettings reads the user settings from the internal endpoints of auth-service,
// authenticated with a service token for audience.
type HTTPUserSettings struct {
	client   *http.Client
	baseURL  string
	audience string
	signer   ITokenSigner
}

func NewHTTPUserSettings(client *http.Client, baseURL, audience string, signer ITokenSigner) *HTTPUserSettings {
	return &HTTPUserSettings{
		client:   client,
		baseURL:  strings.TrimRight(baseURL, "/"),
		audience: audience,
		signer:   signer,
	}
}

type dropboxResponse struct {
	Token  string `json:"token"`
	Folder string `json:"folder"`
}

type fcmTokenResponse struct {
	Token string `json:"token"`
}

func (h *HTTPUserSettings) GetDropboxAccount(ctx context.Context, userId int) (*models.DropboxAccount, error) {
	var res dropboxResponse
	if err := h.get(ctx, fmt.Sprintf("/internal/users/%d/dropbox", userId), &res, models.ErrDropboxNotLinked); err != nil {
		return nil, err
	}
	return models.NewDropboxAccount(res.Token, res.Folder)
}

func (h *HTTPUserSettings) GetFCMToken(ctx context.Context, userId int) (string, error) {
	var res fcmTokenResponse
	if err := h.get(ctx, fmt.Sprintf("/internal/users/%d/fcm-token", userId), &res, models.ErrFCMTokenNotFound); err != nil {
		return "", err
	}
	if res.Token == "" {
		return "", models.ErrFCMTokenNotFound
	}
	return res.Token, nil
}

// get decodes the JSON body of GET path into out, returning notFound on a 404.
func (h *HTTPUserSettings) get(ctx context.Context, path string, out any, notFound error) error {
	token, err := h.signer.Sign(h.audience)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, h.baseURL+path, nil)
	if err != nil {
		return fmt.Errorf("create request %s: %w", path, err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")

	res, err := h.client.Do(req)
	if err != nil {
		return fmt.Errorf("get %s: %w", path, err)
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		return notFound
	}
	if res.StatusCode != http.StatusOK {
		// the body only adds detail to the error, so a failed read is ignored
		body, _ := io.ReadAll(io.LimitReader(res.Body, 1<<10))
		return fmt.Errorf("get %s: status %d: %s", path, res.StatusCode, strings.TrimSpace(string(body)))
	}

	if err := json.NewDecoder(res.Body).Decode(out); err != nil {
		return fmt.Errorf("decode %s: %w", path, err)
	}
	return nil
}
