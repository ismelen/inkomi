package dropbox

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ismelen/inkomi/back/auth-service/internal/domain/models"
	"github.com/ismelen/inkomi/back/auth-service/internal/domain/ports"
)

var _ ports.IDropboxOAuth = (*HTTPDropboxOAuth)(nil)

const APIURL = "https://api.dropboxapi.com"

// HTTPDropboxOAuth refreshes tokens issued to the PKCE flow of the app with appKey, which
// needs no secret.
type HTTPDropboxOAuth struct {
	client *http.Client
	apiURL string
	appKey string
}

func NewHTTPDropboxOAuth(client *http.Client, apiURL, appKey string) *HTTPDropboxOAuth {
	return &HTTPDropboxOAuth{client: client, apiURL: strings.TrimRight(apiURL, "/"), appKey: appKey}
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
}

type accountResponse struct {
	AccountId     string `json:"account_id"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
}

func (h *HTTPDropboxOAuth) Refresh(ctx context.Context, refreshToken string) (models.AccessToken, error) {
	if refreshToken == "" {
		return models.AccessToken{}, models.ErrInvalidExternalToken
	}

	form := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {refreshToken},
		"client_id":     {h.appKey},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, h.apiURL+"/oauth2/token", strings.NewReader(form.Encode()))
	if err != nil {
		return models.AccessToken{}, fmt.Errorf("create dropbox token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	var res tokenResponse
	if err := h.do(req, &res); err != nil {
		return models.AccessToken{}, err
	}
	if res.AccessToken == "" {
		return models.AccessToken{}, fmt.Errorf("dropbox token response without access token")
	}
	return models.AccessToken{
		Token:     res.AccessToken,
		ExpiresAt: time.Now().Add(time.Duration(res.ExpiresIn) * time.Second),
	}, nil
}

func (h *HTTPDropboxOAuth) Account(ctx context.Context, accessToken string) (*models.ExternalAccount, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, h.apiURL+"/2/users/get_current_account", nil)
	if err != nil {
		return nil, fmt.Errorf("create dropbox account request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	var res accountResponse
	if err := h.do(req, &res); err != nil {
		return nil, err
	}
	if res.AccountId == "" {
		return nil, fmt.Errorf("dropbox account response without account id")
	}
	return &models.ExternalAccount{
		Provider:      models.ProviderDropbox,
		Subject:       res.AccountId,
		Email:         res.Email,
		EmailVerified: res.EmailVerified,
	}, nil
}

// do decodes a 200 response into out. Dropbox rejects revoked, expired or foreign tokens with
// a 400 or 401, which is returned as models.ErrInvalidExternalToken.
func (h *HTTPDropboxOAuth) do(req *http.Request, out any) error {
	res, err := h.client.Do(req)
	if err != nil {
		return fmt.Errorf("dropbox %s: %w", req.URL.Path, err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		// the body only adds detail to the error, so a failed read is ignored
		body, _ := io.ReadAll(io.LimitReader(res.Body, 1<<10))
		err := fmt.Errorf("dropbox %s: status %d: %s", req.URL.Path, res.StatusCode, strings.TrimSpace(string(body)))
		if res.StatusCode == http.StatusBadRequest || res.StatusCode == http.StatusUnauthorized {
			return fmt.Errorf("%w: %w", models.ErrInvalidExternalToken, err)
		}
		return err
	}

	if err := json.NewDecoder(res.Body).Decode(out); err != nil {
		return fmt.Errorf("decode dropbox %s: %w", req.URL.Path, err)
	}
	return nil
}
