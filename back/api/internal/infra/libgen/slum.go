package libgen

import (
	"ismelen/inkomi/internal/domain/book"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

var slumURL = "https://open-slum.org/libgen.html"

// getMirrors scrapes open-slum.org/libgen.html and returns all mirrors whose
// current status badge is "up" (case-insensitive). Protected, degraded and
// down mirrors are excluded. If the page cannot be fetched or no UP mirrors
// are found, an empty slice is returned — no hardcoded fallback list.
var getMirrors = func() []book.BooksSource {
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", slumURL, nil)
	if err != nil {
		return nil
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil
	}

	var mirrors []book.BooksSource

	// Each mirror is wrapped in a <div class="card" id="libgen.XX">.
	// Inside it we find:
	//   <a href="https://libgen.XX" class="card-title">  → the mirror URL
	//   <span class="status-badge up|protected|degraded|down">  → current status
	//
	// We only include mirrors whose status-badge class list contains "up"
	// (the badge has two classes: "status-badge" and the status itself).
	doc.Find(".card").Each(func(_ int, card *goquery.Selection) {
		status := strings.ToLower(strings.TrimSpace(card.Find(".status-badge").Last().Text()))
		if status != "up" {
			return
		}

		href, exists := card.Find("a.card-title").Attr("href")
		if !exists || href == "" {
			return
		}

		base := strings.TrimRight(href, "/")

		// Determine mirror type by probing which search endpoint exists.
		// Plus mirrors expose /index.php; classic mirrors expose /search.php.
		// We defer this detection to SourceDiscoverer's health-check, so here
		// we try Plus first (it's the more common modern variant) and fall back
		// to Classic. Both are created and the discoverer will probe with a
		// known MD5 to find which one actually works.
		mirrors = append(mirrors, NewPlusMirror(base))
		mirrors = append(mirrors, NewClassicMirror(base))
	})

	return mirrors
}
