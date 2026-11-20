package libgen

import (
	"context"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/ismelen/inkomi/back/api/internal/modules/library/domain/models"
	"github.com/ismelen/inkomi/back/api/internal/modules/library/shared/strutil"
)

var hrefMD5Re = regexp.MustCompile(`(?i)[0-9a-f]{32}`)

// PlusMirror is a libgen+ site (libgen.la, libgen.bz...): results are scraped from index.php.
type PlusMirror struct {
	mirrorBase
}

func NewPlusMirror(url string) *PlusMirror {
	return &PlusMirror{mirrorBase{url: url}}
}

func (p *PlusMirror) Search(ctx context.Context, query string, field models.SearchField) ([]models.Book, error) {
	params := url.Values{}
	params.Set("req", query)
	params.Set("res", strconv.Itoa(maxResults))
	params.Set("filesuns", "all")
	params.Set("objects[]", "f")
	params.Add("columns[]", plusColumn(field))
	for _, topic := range []string{"l", "c", "f", "a", "m", "r", "s"} {
		params.Add("topics[]", topic)
	}
	params.Set("covers", "on")
	params.Set("curtab", "f")
	params.Set("order", "time_added")
	params.Set("ordermode", "desc")

	doc, err := fetchDocument(ctx, p.url+"/index.php?"+params.Encode())
	if err != nil {
		return nil, err
	}

	books := make([]models.Book, 0)
	doc.Find("#tablelibgen tr").Each(func(i int, row *goquery.Selection) {
		if i == 0 {
			return
		}
		if book, ok := p.parseRow(row.Find("td")); ok {
			books = append(books, book)
		}
	})
	return books, nil
}

func (p *PlusMirror) parseRow(tds *goquery.Selection) (models.Book, bool) {
	if tds.Length() <= 9 {
		return models.Book{}, false
	}

	pages, err := strconv.Atoi(strings.TrimSpace(tds.Eq(6).Text()))
	if err != nil {
		return models.Book{}, false
	}

	book := models.Book{
		Title:     rowTitle(tds.Eq(1)),
		Author:    strings.TrimSpace(tds.Eq(2).Text()),
		Language:  strings.TrimSpace(tds.Eq(5).Text()),
		Pages:     pages,
		Extension: strings.ToLower(strings.TrimSpace(tds.Eq(8).Text())),
	}

	tds.Eq(9).Find("a").EachWithBreak(func(_ int, a *goquery.Selection) bool {
		href, _ := a.Attr("href")
		book.MD5 = strings.ToLower(hrefMD5Re.FindString(href))
		return book.MD5 == ""
	})

	if cover, err := resolveURL(p.url+"/", rowCover(tds.Eq(0))); err == nil && cover != p.url+"/" {
		book.CoverURL = cover
	}
	return book, true
}

func plusColumn(field models.SearchField) string {
	if field == models.SearchFieldAuthor {
		return "a"
	}
	return "t"
}

// rowCover reads the cover of a result, which may be lazy loaded (data-src) behind a placeholder.
func rowCover(td *goquery.Selection) string {
	img := td.Find("img").First()
	for _, attr := range []string{"data-src", "src"} {
		if src, ok := img.Attr(attr); ok && isCoverPath(src) {
			return strings.TrimSpace(src)
		}
	}
	return ""
}

// rowTitle prefers the longest edition link and falls back to the last meaningful text line.
func rowTitle(td *goquery.Selection) string {
	title := ""
	td.Find("a[href*='edition.php']").Each(func(_ int, a *goquery.Selection) {
		if t := strings.TrimSpace(a.Text()); len(t) > len(title) {
			title = t
		}
	})
	if title != "" {
		return title
	}

	lines := strings.Split(strings.TrimSpace(td.Text()), "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line != "" && !strutil.IsNumeric(line) && line != "c" && line != "f" {
			return line
		}
	}
	return ""
}
