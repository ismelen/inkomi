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
	LibgenMirrorBase
	url string
}

func NewPlusMirror(url string) PlusMirror {
	return PlusMirror{url: url}
}

func (p PlusMirror) GetURL() string {
	return p.url
}

func (p PlusMirror) Download(req book.LibgenDownloadRequest) (*book.LibgenDownloadResult, error) {
	return p.LibgenMirrorBase.Download(req)
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
