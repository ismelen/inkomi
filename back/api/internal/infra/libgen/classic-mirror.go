package libgen

import (
	"encoding/json"
	"fmt"
	"io"
	"ismelen/inkomi/internal/domain"
	"ismelen/inkomi/internal/infra/helpers"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type ClassicMirror struct {
	LibgenMirrorBase
	url string
}

func NewClassicMirror(url string) ClassicMirror {
	return ClassicMirror{url: url}
}

func (c ClassicMirror) GetURL() string {
	return c.url
}

func (c ClassicMirror) Search(query string) ([]domain.Book, error) {
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
	params.Set("fields", "id,title,author,year,publisher,pages,language,filesize,extension,md5")
	jsonURL := c.url + "/json.php?" + params.Encode()

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

	var books []domain.Book
	if err := json.Unmarshal(body, &books); err != nil {
		preview := string(body)
		if len(preview) > 300 {
			preview = preview[:300]
		}
		return nil, fmt.Errorf("JSON inválido: %w\n— %s", err, preview)
	}

	for i := range books {
		books[i].CoverURL = c.buildCoverURL(books[i])
		books[i].DownloadURL = "https://library.lol/main/" + strings.ToLower(books[i].MD5)
	}

	if len(books) == 0 {
		return nil, fmt.Errorf("No books were found")
	}

	return books, nil
}

func (c ClassicMirror) buildCoverURL(b domain.Book) string {
	if b.MD5 == "" {
		return ""
	}
	idNum, err := strconv.Atoi(b.ID)
	if err != nil {
		return fmt.Sprintf("%s/covers/%s.jpg", c.url, strings.ToLower(b.MD5))
	}
	folder := (idNum / 1000) * 1000
	return fmt.Sprintf("%s/covers/%d/%s.jpg", c.url, folder, strings.ToLower(b.MD5))
}

func (c ClassicMirror) getIds(query string) ([]string, error) {
	params := url.Values{}
	params.Set("req", query)
	params.Set("res", "25")
	params.Set("view", "simple")
	params.Set("phrase", "1")
	params.Set("column", "def")
	searchURL := c.url + "/search.php?" + params.Encode()

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
		if idText != "" && helpers.IsNumeric(idText) {
			if !slices.Contains(ids, idText) {
				ids = append(ids, idText)
			}
		}
	})
	return ids, nil
}
