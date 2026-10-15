package libgen

import (
	"context"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/ismelen/inkomi/back/library-service/internal/domain/models"
)

const maxAdsPageSize = 1 << 20

var (
	adsTitleRe    = regexp.MustCompile(`(?i)Title:\s*(.*?)<br>`)
	adsExtRe      = regexp.MustCompile(`(?i)Extension:\s*([^,<]+)`)
	adsDownloadRe = regexp.MustCompile(`(?i)href=["']([^"']+)["'][^>]*><h2>GET</h2>`)
)

// mirrorBase holds what every libgen flavour shares: the ads.php download page.
type mirrorBase struct {
	url string
}

func (m *mirrorBase) URL() string { return m.url }

func (m *mirrorBase) Download(ctx context.Context, md5 models.MD5) (*models.BookDownload, error) {
	info, err := m.adsInfo(ctx, md5)
	if err != nil {
		return nil, err
	}

	resp, err := get(ctx, downloadClient, info.downloadURL)
	if err != nil {
		return nil, fmt.Errorf("download %q: %w", info.title, err)
	}

	return &models.BookDownload{
		Stream: resp.Body,
		Title:  info.title,
		Ext:    info.ext,
	}, nil
}

type adsInfo struct {
	title, ext, downloadURL string
}

func (m *mirrorBase) adsInfo(ctx context.Context, md5 models.MD5) (*adsInfo, error) {
	pageURL := m.url + "/ads.php?md5=" + string(md5)
	resp, err := get(ctx, pageClient, pageURL)
	if err != nil {
		var statusErr *statusError
		if errors.As(err, &statusErr) && statusErr.code == http.StatusNotFound {
			return nil, fmt.Errorf("%s in %s: %w", md5, m.url, models.ErrBookNotFound)
		}
		return nil, err
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxAdsPageSize))
	resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", pageURL, err)
	}

	page := string(body)
	info := &adsInfo{}
	if match := adsTitleRe.FindStringSubmatch(page); match != nil {
		info.title = strings.TrimSpace(html.UnescapeString(match[1]))
	}
	if match := adsExtRe.FindStringSubmatch(page); match != nil {
		info.ext = strings.ToLower(strings.TrimSpace(match[1]))
	}
	if match := adsDownloadRe.FindStringSubmatch(page); match != nil {
		info.downloadURL, err = resolveURL(pageURL, html.UnescapeString(strings.TrimSpace(match[1])))
		if err != nil {
			return nil, err
		}
	}

	if info.title == "" || info.downloadURL == "" {
		return nil, fmt.Errorf("%s in %s: %w", md5, m.url, models.ErrBookNotFound)
	}
	return info, nil
}

func resolveURL(base, ref string) (string, error) {
	baseURL, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("parse %s: %w", base, err)
	}
	refURL, err := url.Parse(ref)
	if err != nil {
		return "", fmt.Errorf("parse %s: %w", ref, err)
	}
	return baseURL.ResolveReference(refURL).String(), nil
}
