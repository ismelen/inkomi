package libgen

import (
	"crypto/tls"
	"fmt"
	"ismelen/inkomi/internal/domain/book"
	"ismelen/inkomi/internal/shared/strutil"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// metadataSuffixes are the libgen ads.php field labels that may appear
// concatenated after the real title (e.g. "My Book Author(s): John Doe …").
var metadataSuffixes = []string{
	" Author(s):",
	" Authors:",
	" Series:",
	" Publisher:",
	" Year:",
	" ISBN:",
	" Pages:",
	" Language:",
	" Edition:",
	" Volume:",
}

// stripLibgenMetadata removes any trailing libgen metadata that ads.php
// sometimes concatenates into the title <b> element.
func stripLibgenMetadata(title string) string {
	for _, suffix := range metadataSuffixes {
		if idx := strings.Index(title, suffix); idx != -1 {
			title = title[:idx]
		}
	}
	return strings.TrimSpace(title)
}

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
	baseTransport = func() *http.Transport {
		t := http.DefaultTransport.(*http.Transport).Clone()
		t.ForceAttemptHTTP2 = false
		t.DisableCompression = true
		t.TLSClientConfig = &tls.Config{
			NextProtos: []string{"http/1.1"},
		}
		return t
	}()

	downloadTransport = func() *http.Transport {
		t := baseTransport.Clone()
		t.ResponseHeaderTimeout = 15 * time.Second
		return t
	}()

	httpClient     = &http.Client{Timeout: 20 * time.Second, Transport: baseTransport}
	downloadClient = &http.Client{Timeout: 10 * time.Minute, Transport: downloadTransport}
)

func (m MirrorBase) FetchURL(rawURL string, isDownload bool) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,*/*;q=0.8")
	req.Header.Set("Accept-Language", "es-ES,es;q=0.9,en;q=0.8")
	req.Header.Set("Accept-Encoding", "identity")
	req.Header.Set("Referer", m.Url+"/")

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
		resp.Body.Close()
		return nil, fmt.Errorf("'%s' download failed", data.title)
	}

	safeTitle := strutil.SanitizeFilename(stripLibgenMetadata(data.title))
	filename := safeTitle + "." + data.extension

	return &book.LibgenDownload{
		Stream:        resp.Body,
		ContentType:   resp.Header.Get("Content-Type"),
		ContentLength: resp.ContentLength,
		Filename:      filename,
		Title:         data.title,
		Ext:           data.extension,
	}, nil
}

type basicBook struct {
	title, downloadUrl, extension, md5 string
}

// GetBasicBookFromMD5 fetches the ads.php page for the given MD5 and parses
// title, extension and download URL using goquery instead of fragile regexes.
// The ads.php page layout:
//
//	<td>Title: <b>…</b></td>          or    Title: …<br>
//	<td>Extension: <b>epub</b></td>   or    Extension: epub,  Size: …<br>
//	<a href="…"><h2>GET</h2></a>       — the direct download link
func (m MirrorBase) GetBasicBookFromMD5(md5 string) (*basicBook, error) {
	adsURL := m.Url + "/ads.php?md5=" + md5
	doc, err := m.Fetch(adsURL)
	if err != nil {
		return nil, err
	}

	bk := &basicBook{md5: md5}

	// --- Title ---
	// Look for a <td> or any element whose text starts with "Title:"
	doc.Find("td, li, p").EachWithBreak(func(_ int, s *goquery.Selection) bool {
		text := strings.TrimSpace(s.Text())
		if strings.HasPrefix(strings.ToLower(text), "title:") {
			// Prefer the text inside a <b> or <a> child if present
			if child := s.Find("b, a").First(); child.Length() > 0 {
				bk.title = stripLibgenMetadata(strings.TrimSpace(child.Text()))
			} else {
				after, _ := strings.CutPrefix(strings.ToLower(text), "title:")
				bk.title = stripLibgenMetadata(strings.TrimSpace(text[len(text)-len(after):]))
			}
			return bk.title == ""
		}
		return true
	})

	// --- Extension ---
	doc.Find("td, li, p").EachWithBreak(func(_ int, s *goquery.Selection) bool {
		text := strings.TrimSpace(s.Text())
		lower := strings.ToLower(text)
		if strings.HasPrefix(lower, "extension:") {
			if child := s.Find("b").First(); child.Length() > 0 {
				bk.extension = strings.ToLower(strings.TrimSpace(child.Text()))
			} else {
				after, _ := strings.CutPrefix(lower, "extension:")
				// Extension field may be "epub, Size: 1.2 MB" — take first token
				bk.extension = strings.ToLower(strings.TrimSpace(strings.SplitN(after, ",", 2)[0]))
			}
			return bk.extension == ""
		}
		return true
	})

	// --- Download URL ---
	// Find any <a> that wraps an <h2> whose text is "GET" (case-insensitive),
	// or whose own text is "GET". This is more robust than a single regex.
	doc.Find("a").EachWithBreak(func(_ int, a *goquery.Selection) bool {
		h2 := a.Find("h2")
		linkText := strings.ToUpper(strings.TrimSpace(a.Text()))
		h2Text := strings.ToUpper(strings.TrimSpace(h2.Text()))

		if h2Text != "GET" && linkText != "GET" {
			return true
		}

		href, exists := a.Attr("href")
		if !exists || href == "" {
			return true
		}

		href = strings.TrimSpace(href)
		switch {
		case strings.HasPrefix(href, "http"):
			bk.downloadUrl = href
		case strings.HasPrefix(href, "/"):
			bk.downloadUrl = m.Url + href
		default:
			bk.downloadUrl = m.Url + "/" + href
		}
		return false
	})

	if bk.title == "" || bk.downloadUrl == "" {
		return nil, fmt.Errorf("no book data from %s", m.Url)
	}

	if bk.extension == "" {
		bk.extension = "epub"
	}

	return bk, nil
}
