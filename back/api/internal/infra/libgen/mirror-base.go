package libgen

import (
	"fmt"
	"ismelen/inkomi/internal/domain"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type LibgenMirrorBase struct{}

func (l LibgenMirrorBase) Fetch(url string) (*goquery.Document, error) {
	resp, err := l.FetchURL(url, false)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d en búsqueda", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error parseando HTML: %w", err)
	}

	return doc, nil
}

var (
	httpClient     = &http.Client{Timeout: 20 * time.Second}
	downloadClient = &http.Client{Timeout: 10 * time.Minute}
)

func (l LibgenMirrorBase) FetchURL(rawURL string, isDownload bool) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,*/*;q=0.8")
	req.Header.Set("Accept-Language", "es-ES,es;q=0.9,en;q=0.8")

	if isDownload {
		return downloadClient.Do(req)
	}
	return httpClient.Do(req)
}

func (l LibgenMirrorBase) Download(req domain.LibgenDownloadRequestDTO) (*domain.LibgenDownloadResultDTO, error) {
	dlURL, err := l.resolveDownloadLink(req.DownloadURL)
	if err != nil || dlURL == "" {
		dlURL = req.DownloadURL
	}

	resp, err := l.FetchURL(dlURL, true)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("HTTP %d al descargar", resp.StatusCode)
	}

	ext := req.Extension
	if ext == "" {
		ext = "epub"
	}
	filename := sanitizeFilename(req.Title) + "." + strings.ToLower(ext)
	filename = filepath.Clean(filename)

	return &domain.LibgenDownloadResultDTO{
		Stream:        resp.Body,
		ContentType:   resp.Header.Get("Content-Type"),
		ContentLength: resp.ContentLength,
		Filename:      filename,
	}, nil
}

func (l LibgenMirrorBase) resolveDownloadLink(pageURL string) (string, error) {
	resp, err := l.FetchURL(pageURL, false)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", err
	}

	var link string
	for _, sel := range []string{"#download a", "a[href^='get.php']", "a[href*='cloudflare-ipfs']", "a[href*='ipfs.io']"} {
		if link != "" {
			break
		}
		doc.Find(sel).Each(func(_ int, s *goquery.Selection) {
			if link != "" {
				return
			}
			if href, ok := s.Attr("href"); ok {
				if !strings.HasPrefix(href, "http") {
					u, _ := url.Parse(pageURL)
					href = fmt.Sprintf("%s://%s/%s", u.Scheme, u.Host, strings.TrimPrefix(href, "/"))
				}
				link = href
			}
		})
	}
	return link, nil
}

func sanitizeFilename(s string) string {
	rep := strings.NewReplacer(
		"/", "-", "\\", "-", ":", "-", "*", "",
		"?", "", "\"", "", "<", "", ">", "", "|", "",
		"\n", " ", "\r", "",
	)
	result := strings.TrimSpace(rep.Replace(s))
	if len(result) > 120 {
		result = result[:120]
	}
	if result == "" {
		result = "libro"
	}
	return result
}
