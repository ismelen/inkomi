package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/ismelen/inkomi/back/release-service/internal/domain/models"
	"github.com/ismelen/inkomi/back/release-service/internal/domain/ports"
)

const requestTimeout = 10 * time.Second

var _ ports.IReleaseSource = (*GitHubReleaseSource)(nil)

// GitHubReleaseSource reads the latest non draft, non prerelease release of a repository.
type GitHubReleaseSource struct {
	client    *http.Client
	apiURL    string
	repo      string
	token     string
	assetName string
}

func NewGitHubReleaseSource(apiURL, repo, token, assetName string) *GitHubReleaseSource {
	return &GitHubReleaseSource{
		client:    &http.Client{Timeout: requestTimeout},
		apiURL:    strings.TrimRight(apiURL, "/"),
		repo:      repo,
		token:     token,
		assetName: assetName,
	}
}

type releaseResponse struct {
	TagName     string          `json:"tag_name"`
	Body        string          `json:"body"`
	PublishedAt time.Time       `json:"published_at"`
	Assets      []assetResponse `json:"assets"`
}

type assetResponse struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
	// Digest is "sha256:<hex>" on assets uploaded since mid 2025.
	Digest string `json:"digest"`
}

func (g *GitHubReleaseSource) Latest(ctx context.Context) (*models.Release, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, g.apiURL+"/repos/"+g.repo+"/releases/latest", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("User-Agent", "inkomi-release-service")
	if g.token != "" {
		req.Header.Set("Authorization", "Bearer "+g.token)
	}

	res, err := g.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", models.ErrSourceUnavailable, err)
	}
	defer res.Body.Close()

	switch {
	case res.StatusCode == http.StatusNotFound:
		return nil, models.ErrReleaseNotFound
	case res.StatusCode != http.StatusOK:
		body, _ := io.ReadAll(io.LimitReader(res.Body, 512))
		return nil, fmt.Errorf("%w: github %d: %s", models.ErrSourceUnavailable, res.StatusCode, body)
	}

	var payload releaseResponse
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode github release: %w", err)
	}

	return g.toRelease(payload)
}

func (g *GitHubReleaseSource) toRelease(payload releaseResponse) (*models.Release, error) {
	version, err := models.ParseVersion(payload.TagName)
	if err != nil {
		return nil, err
	}

	for _, asset := range payload.Assets {
		if asset.Name != g.assetName {
			continue
		}
		return &models.Release{
			Version:     version,
			Notes:       payload.Body,
			PublishedAt: payload.PublishedAt,
			APK: models.ReleaseAsset{
				Name:   asset.Name,
				URL:    asset.BrowserDownloadURL,
				Size:   asset.Size,
				SHA256: strings.TrimPrefix(asset.Digest, "sha256:"),
			},
		}, nil
	}

	return nil, models.ErrAPKNotFound
}
