package libgen

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/ismelen/inkomi/back/api/internal/modules/library/domain/models"
	"github.com/ismelen/inkomi/back/api/internal/modules/library/domain/ports"
)

const maxCoverSize = 5 << 20

var coverExts = map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true}

// LibgenCoverSource downloads the cover images of the libgen mirrors. It only fetches images
// from libgen hosts, so the cover route can't be used to reach anything else.
type LibgenCoverSource struct{}

var _ ports.ICoverSource = (*LibgenCoverSource)(nil)

func NewLibgenCoverSource() *LibgenCoverSource {
	return &LibgenCoverSource{}
}

func (s *LibgenCoverSource) Cover(ctx context.Context, rawURL string) (*models.Cover, error) {
	if !IsCoverURL(rawURL) {
		return nil, models.ErrInvalidCoverURL
	}
	return fetchCover(ctx, rawURL)
}

// fetchCover downloads the image at rawURL as a page of its own mirror would. The mirrors
// protect their covers from hotlinking: without a Referer of the same site they answer
// 200 image/jpeg with an empty body.
func fetchCover(ctx context.Context, rawURL string) (*models.Cover, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", rawURL, models.ErrInvalidCoverURL)
	}
	headers := http.Header{}
	headers.Set("Referer", u.Scheme+"://"+u.Host+"/")
	headers.Set("Accept", "image/avif,image/webp,image/apng,image/*,*/*;q=0.8")

	resp, err := getWithHeaders(ctx, pageClient, rawURL, headers)
	if err != nil {
		var statusErr *statusError
		if errors.As(err, &statusErr) && statusErr.code == http.StatusNotFound {
			return nil, fmt.Errorf("%s: %w", rawURL, models.ErrInvalidCoverURL)
		}
		return nil, err
	}
	contentType := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		resp.Body.Close()
		return nil, fmt.Errorf("%s is %q, not an image: %w", rawURL, contentType, models.ErrInvalidCoverURL)
	}
	if resp.ContentLength == 0 {
		resp.Body.Close()
		return nil, fmt.Errorf("%s is an empty image: %w", rawURL, models.ErrInvalidCoverURL)
	}
	return &models.Cover{
		Body:        limitedBody{Reader: io.LimitReader(resp.Body, maxCoverSize), Closer: resp.Body},
		ContentType: contentType,
	}, nil
}

// IsCoverURL reports whether rawURL is an image on a libgen host.
func IsCoverURL(rawURL string) bool {
	u, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || (u.Scheme != "https" && u.Scheme != "http") || u.User != nil {
		return false
	}
	host := strings.ToLower(u.Hostname())
	if !strings.HasPrefix(host, "libgen.") || strings.Count(host, ".") != 1 {
		return false
	}
	return isCoverPath(u.Path)
}

// isCoverPath reports whether p looks like a cover image rather than a placeholder.
func isCoverPath(p string) bool {
	p = strings.ToLower(strings.TrimSpace(p))
	if u, err := url.Parse(p); err == nil {
		p = u.Path
	}
	name := path.Base(p)
	return coverExts[path.Ext(name)] && !strings.Contains(name, "blank") && !strings.Contains(name, "nocover")
}

type limitedBody struct {
	io.Reader
	io.Closer
}
