package libgen

import (
	"fmt"
	"ismelen/inkomi/internal/domain/book"
	"ismelen/inkomi/internal/shared/strutil"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type PlusMirror struct {
	MirrorBase
	url string
}

func NewPlusMirror(url string) PlusMirror {
	return PlusMirror{url: url}
}

func (p PlusMirror) GetURL() string {
	return p.url
}

func (p PlusMirror) Search(query string) ([]book.Book, error) {
	params := url.Values{}
	params.Set("req", query)
	params.Set("res", "25")
	params.Set("filesuns", "all")
	for _, obj := range []string{"f", "e", "s", "a", "p", "w"} {
		params.Add("objects[]", obj)
	}
	for _, topic := range []string{"l", "c", "f", "a", "m", "r", "s"} {
		params.Add("topics[]", topic)
	}
	searchURL := p.url + "/index.php?" + params.Encode()

	doc, err := p.Fetch(searchURL)
	if err != nil {
		return nil, fmt.Errorf("error parseando HTML: %w", err)
	}

	idRe := regexp.MustCompile(`[?&]id=(\d+)`)
	md5Re := regexp.MustCompile(`(?i)[0-9a-f]{32}`)

	var books []book.Book

	doc.Find("#tablelibgen tr").Each(func(i int, row *goquery.Selection) {
		if i == 0 {
			return
		}
		tds := row.Find("td")
		if tds.Length() < 8 {
			return
		}

		var editionID string
		isFiction := false
		tds.Eq(0).Find("a[href*='edition.php']").Each(func(_ int, a *goquery.Selection) {
			if editionID != "" {
				return
			}
			href, _ := a.Attr("href")
			matches := idRe.FindStringSubmatch(href)
			if len(matches) >= 2 {
				editionID = matches[1]
			}
		})
		if editionID == "" {
			return
		}

		td0Text := strings.TrimSpace(tds.Eq(0).Text())
		words := strings.Fields(td0Text)
		mainID := ""
		if len(words) > 0 {
			lastWord := words[len(words)-1]
			if strutil.IsNumeric(lastWord) {
				mainID = lastWord
			}
			for j := len(words) - 1; j >= 0 && j >= len(words)-3; j-- {
				if words[j] == "f" {
					isFiction = true
					break
				}
			}
		}
		if mainID == "" {
			mainID = editionID
		}

		title := ""
		tds.Eq(0).Find("a[href*='edition.php']").Each(func(_ int, a *goquery.Selection) {
			t := strings.TrimSpace(a.Text())
			if len(t) > len(title) {
				title = t
			}
		})
		if title == "" {
			lines := strings.Split(td0Text, "\n")
			for j := len(lines) - 1; j >= 0; j-- {
				line := strings.TrimSpace(lines[j])
				if line != "" && !strutil.IsNumeric(line) && line != "c" && line != "f" {
					title = line
					break
				}
			}
		}

		md5 := ""
		tds.Eq(8).Find("a").Each(func(_ int, a *goquery.Selection) {
			if md5 != "" {
				return
			}
			href, _ := a.Attr("href")
			matches := md5Re.FindString(strings.ToLower(href))
			if len(matches) == 32 {
				md5 = matches
			}
		})

		author := strings.TrimSpace(tds.Eq(1).Text())
		lang := strings.TrimSpace(tds.Eq(4).Text())
		pages := strings.TrimSpace(tds.Eq(5).Text())
		ext := strings.ToLower(strings.TrimSpace(tds.Eq(7).Text()))

		downloadUrl := p.url + "/ads.php?md5=" + strings.ToLower(md5)
		if md5 == "" {
			downloadUrl = p.url + "/edition.php?id=" + mainID
		}

		book := book.Book{
			Title:       title,
			Author:      author,
			Pages:       pages,
			Language:    lang,
			Extension:   ext,
			MD5:         md5,
			CoverURL:    p.buildCoverURL(mainID, md5, isFiction),
			DownloadURL: downloadUrl,
		}

		books = append(books, book)
	})

	if len(books) == 0 {
		return nil, fmt.Errorf("No books were found")
	}

	return books, nil
}

func (p PlusMirror) buildCoverURL(id string, md5 string, isFiction bool) string {
	if md5 == "" {
		return ""
	}
	idNum, err := strconv.Atoi(id)
	if err != nil {
		return fmt.Sprintf("%s/covers/%s.jpg", p.url, strings.ToLower(md5))
	}
	folder := (idNum / 1000) * 1000
	coverDir := "covers"
	if isFiction {
		coverDir = "fictioncovers"
	}
	return fmt.Sprintf("%s/%s/%d/%s.jpg", p.url, coverDir, folder, strings.ToLower(md5))
}


package libgen

import (
	"context"
	"fmt"
	"libgen-search/domain"
	"net/http"
	"net/url"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// md5Re se compila una sola vez a nivel de paquete para evitar el coste
// de recompilación en cada llamada a searchPlusWithMetadata.
var md5Re = regexp.MustCompile(`(?i)[0-9a-f]{32}`)

func searchPlusWithMetadata(ctx context.Context, query string, m mirror) ([]domain.Book, error) {
	params := url.Values{}
	params.Set("req", query)
	params.Set("res", "25")
	params.Set("filesuns", "all")
	params.Set("objects[]", "f")
	for _, obj := range []string{"t", "a", "s", "y", "p", "i", "f"} {
		params.Add("columns[]", obj)
	}
	for _, topic := range []string{"l", "c", "f", "a", "m", "r", "s"} {
		params.Add("topics[]", topic)
	}
	params.Set("covers", "on")
	params.Set("curtab", "f")
	params.Set("order", "time_added")
	params.Set("ordermode", "desc")

	searchURL := m.base + "/index.php?" + params.Encode()

	resp, err := fetchURL(ctx, searchURL, false)
	if err != nil {
		return nil, fmt.Errorf("error de red: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d en búsqueda", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error parseando HTML: %w", err)
	}

	books := make([]domain.Book, 0)
	doc.Find("#tablelibgen tr").Each(func(i int, row *goquery.Selection) {
		if i == 0 {
			return
		}
		tds := row.Find("td")
		if tds.Length() <= 9 {
			return
		}

		pagesStr := strings.TrimSpace(tds.Eq(6).Text())
		pages, err := strconv.Atoi(pagesStr)
		if err != nil {
			return
		}
		if pages == 0 {
			return
		}

		title := ""
		tds.Eq(1).Find("a[href*='edition.php']").EachWithBreak(func(_ int, a *goquery.Selection) bool {
			t := strings.TrimSpace(a.Text())
			if len(t) > len(title) {
				title = t
				return false
			}
			return true
		})

		if title == "" {
			td0Text := strings.TrimSpace(tds.Eq(1).Text())
			lines := strings.Split(td0Text, "\n")
			for j := len(lines) - 1; j >= 0; j-- {
				line := strings.TrimSpace(lines[j])
				if line != "" && !isNumeric(line) && line != "c" && line != "f" {
					title = line
					break
				}
			}
		}

		md5 := ""
		tds.Eq(9).Find("a").EachWithBreak(func(_ int, a *goquery.Selection) bool {
			href, _ := a.Attr("href")
			if m := md5Re.FindString(strings.ToLower(href)); len(m) == 32 {
				md5 = m
				return false // MD5 encontrado, interrumpir el loop
			}
			return true
		})

		coverUrl, ok := tds.Eq(0).Find("img").Attr("src")
		if ok {
			coverUrl = m.base + coverUrl
		}

		author := strings.TrimSpace(tds.Eq(2).Text())
		lang := strings.TrimSpace(tds.Eq(5).Text())
		ext := strings.ToLower(strings.TrimSpace(tds.Eq(8).Text()))

		b := domain.Book{
			Title:     title,
			Author:    author,
			Language:  lang,
			Extension: ext,
			MD5:       md5,
			CoverURL:  coverUrl,
			Pages:     pages,
		}
		books = append(books, b)
	})

	slices.SortFunc(books, func(a, b domain.Book) int {
		return a.Pages - b.Pages
	})

	return books, nil
}
