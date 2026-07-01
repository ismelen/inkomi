package libgen

import (
	"fmt"
	"ismelen/inkomi/internal/domain/book"
	"ismelen/inkomi/internal/shared/strutil"
	"net/http"
	"net/url"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type PlusMirror struct {
	MirrorBase
}

func NewPlusMirror(url string) PlusMirror {
	return PlusMirror{MirrorBase{url}}
}

var md5Re = regexp.MustCompile(`(?i)[0-9a-f]{32}`)

func (p PlusMirror) Search(query string) ([]book.Book, error) {
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

	searchURL := p.Url + "/index.php?" + params.Encode()

	resp, err := p.FetchURL(searchURL, false)
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

	books := make([]book.Book, 0)
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
				if line != "" && !strutil.IsNumeric(line) && line != "c" && line != "f" {
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
			coverUrl = p.Url + coverUrl
		}

		author := strings.TrimSpace(tds.Eq(2).Text())
		lang := strings.TrimSpace(tds.Eq(5).Text())
		ext := strings.ToLower(strings.TrimSpace(tds.Eq(8).Text()))

		b := book.Book{
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

	slices.SortFunc(books, func(a, b book.Book) int {
		return a.Pages - b.Pages
	})

	return books, nil
}
