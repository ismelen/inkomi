package libgen

import (
	"encoding/json"
	"fmt"
	"io"
	"ismelen/inkomi/internal/domain/book"
	"ismelen/inkomi/internal/shared/strutil"
	"net/http"
	"net/url"
	"slices"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type ClassicMirror struct {
	MirrorBase
}

func NewClassicMirror(url string) ClassicMirror {
	return ClassicMirror{MirrorBase{url}}
}

func (c ClassicMirror) Search(query string) ([]book.Book, error) {
	ids, err := c.getIds(query)
	if err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return nil, fmt.Errorf("sin resultados")
	}

	if len(ids) > 25 {
		ids = ids[:25]
	}

	params := url.Values{}
	params.Set("ids", strings.Join(ids, ","))
	params.Set("fields", "id,title,author,pages,language,extension,md5")
	jsonURL := c.Url + "/json.php?" + params.Encode()

	resp, err := c.FetchURL(jsonURL, false)
	if err != nil {
		return nil, fmt.Errorf("error de red: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d al obtener metadatos", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var books []book.Book
	if err := json.Unmarshal(body, &books); err != nil {
		return nil, fmt.Errorf("JSON inválido")
	}

	if len(books) == 0 {
		return nil, fmt.Errorf("No books were found")
	}

	return books, nil
}

func (c ClassicMirror) getIds(query string) ([]string, error) {
	params := url.Values{}
	params.Set("req", query)
	params.Set("res", "25")
	params.Set("view", "simple")
	params.Set("phrase", "1")
	params.Set("column", "def")
	searchURL := c.Url + "/search.php?" + params.Encode()

	doc, err := c.Fetch(searchURL)
	if err != nil {
		return nil, fmt.Errorf("error parseando HTML: %w", err)
	}

	var ids []string
	doc.Find("table.c tr").Each(func(i int, s *goquery.Selection) {
		if i == 0 {
			return
		}
		idText := strings.TrimSpace(s.Find("td:nth-child(1)").Text())
		if idText != "" && strutil.IsNumeric(idText) {
			if !slices.Contains(ids, idText) {
				ids = append(ids, idText)
			}
		}
	})
	return ids, nil
}
