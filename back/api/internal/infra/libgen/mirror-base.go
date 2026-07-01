package libgen

import (
	"fmt"
	"io"
	"ismelen/inkomi/internal/domain/book"
	"ismelen/inkomi/internal/shared/strutil"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type MirrorBase struct {
	Url string
}

func (m MirrorBase) Fetch(url string) (*goquery.Document, error) {
	resp, err := m.FetchURL(url, false)
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

func (m MirrorBase) FetchURL(rawURL string, isDownload bool) (*http.Response, error) {
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

func (m MirrorBase) Search(query string) ([]book.Book, error) { return nil, nil }

func (m MirrorBase) GetURL() string { return m.Url }

func (m MirrorBase) Download(md5 string) (*book.LibgenDownload, error) {
	data, err := m.GetBasicBookFromMD5(md5)
	if err != nil {
		return nil, err
	}

	resp, err := m.FetchURL(data.downloadUrl, true)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("'%s' download failed", data.title)
	}

	filename := filepath.Clean(strutil.NormalizeString(data.title) + "." + data.extension)

	return &book.LibgenDownload{
		Stream:        resp.Body,
		ContentType:   resp.Header.Get("Content-Type"),
		ContentLength: resp.ContentLength,
		Filename:      filename,
	}, nil
}

type basicBook struct {
	title, downloadUrl, extension, md5 string
}

var titleRe = regexp.MustCompile(`(?i)Title:\s*(.*?)<br>`)
var extRe = regexp.MustCompile(`(?i)Extension:\s*([^,]+).*?Size:\s*(.*?)<br>`)
var downloadRe = regexp.MustCompile(`(?i)href=["']([^"']+)["'][^>]*><h2>GET</h2>`)

func (m MirrorBase) GetBasicBookFromMD5(md5 string) (*basicBook, error) {
	url := m.Url + "/ads.php?md5=" + md5
	resp, err := m.FetchURL(url, false)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("HTTP %d in %s", resp.StatusCode, m.Url)
	}

	b, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, err
	}

	html := string(b)

	if !strings.Contains(html, "Title:") && !strings.Contains(html, "<h2>GET</h2>") {
		return nil, fmt.Errorf("no book data from %s", m.Url)
	}

	book := &basicBook{md5: md5}

	if match := titleRe.FindStringSubmatch(html); len(match) > 1 {
		book.title = strings.TrimSpace(match[1])
	}
	if match := extRe.FindStringSubmatch(html); len(match) > 2 {
		book.extension = strings.TrimSpace(match[1])
	}
	if match := downloadRe.FindStringSubmatch(html); len(match) > 1 {
		href := strings.TrimSpace(match[1])
		if strings.HasPrefix(href, "http") {
			book.downloadUrl = href
		} else if strings.HasPrefix(href, "/") {
			book.downloadUrl = m.Url + href
		} else {
			book.downloadUrl = m.Url + "/" + href
		}
	}

	if book.title == "" || book.downloadUrl == "" {
		return nil, fmt.Errorf("no book data from %s", m.Url)
	}

	if book.extension == "" {
		book.extension = "epub"
	}

	return book, nil
}

// func (l MirrorBase) Download(req book.LibgenDownloadRequest) (*book.LibgenDownloadResult, error) {
// 	dlURL, err := l.resolveDownloadLink(req.DownloadURL)
// 	if err != nil || dlURL == "" {
// 		dlURL = req.DownloadURL
// 	}

// 	resp, err := l.FetchURL(dlURL, true)
// 	if err != nil {
// 		return nil, err
// 	}

// 	if resp.StatusCode != http.StatusOK {
// 		resp.Body.Close()
// 		return nil, fmt.Errorf("HTTP %d al descargar", resp.StatusCode)
// 	}

// 	ext := req.Extension
// 	if ext == "" {
// 		ext = "epub"
// 	}
// 	filename := sanitizeFilename(req.Title) + "." + strings.ToLower(ext)
// 	filename = filepath.Clean(filename)

// 	return &book.LibgenDownloadResult{
// 		Stream:        resp.Body,
// 		ContentType:   resp.Header.Get("Content-Type"),
// 		ContentLength: resp.ContentLength,
// 		Filename:      filename,
// 	}, nil
// }
