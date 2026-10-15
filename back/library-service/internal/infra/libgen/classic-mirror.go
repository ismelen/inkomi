package libgen

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"slices"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/ismelen/inkomi/back/library-service/internal/domain/models"
	"github.com/ismelen/inkomi/back/library-service/internal/shared/strutil"
)

// ClassicMirror is a classic libgen site (libgen.is, libgen.rs...): ids are scraped from
// search.php and their metadata read from json.php.
type ClassicMirror struct {
	mirrorBase
}

func NewClassicMirror(url string) *ClassicMirror {
	return &ClassicMirror{mirrorBase{url: url}}
}

type classicBook struct {
	Title     string `json:"title"`
	Author    string `json:"author"`
	Pages     string `json:"pages"`
	Language  string `json:"language"`
	Extension string `json:"extension"`
	MD5       string `json:"md5"`
}

func (c *ClassicMirror) Search(ctx context.Context, query string) ([]models.Book, error) {
	ids, err := c.searchIds(ctx, query)
	if err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return []models.Book{}, nil
	}

	params := url.Values{}
	params.Set("ids", strings.Join(ids, ","))
	params.Set("fields", "id,title,author,pages,language,extension,md5")
	jsonURL := c.url + "/json.php?" + params.Encode()

	resp, err := get(ctx, pageClient, jsonURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var raw []classicBook
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("decode %s: %w", jsonURL, err)
	}

	books := make([]models.Book, 0, len(raw))
	for _, b := range raw {
		pages, err := strconv.Atoi(strings.TrimSpace(b.Pages))
		if err != nil {
			pages = 0
		}
		books = append(books, models.Book{
			Title:     strings.TrimSpace(b.Title),
			Author:    strings.TrimSpace(b.Author),
			Pages:     pages,
			Language:  strings.TrimSpace(b.Language),
			Extension: strings.ToLower(strings.TrimSpace(b.Extension)),
			MD5:       strings.ToLower(b.MD5),
		})
	}
	return books, nil
}

func (c *ClassicMirror) searchIds(ctx context.Context, query string) ([]string, error) {
	params := url.Values{}
	params.Set("req", query)
	params.Set("res", strconv.Itoa(maxResults))
	params.Set("view", "simple")
	params.Set("phrase", "1")
	params.Set("column", "def")

	doc, err := fetchDocument(ctx, c.url+"/search.php?"+params.Encode())
	if err != nil {
		return nil, err
	}

	var ids []string
	doc.Find("table.c tr").Each(func(i int, row *goquery.Selection) {
		if i == 0 || len(ids) >= maxResults {
			return
		}
		id := strings.TrimSpace(row.Find("td:nth-child(1)").Text())
		if strutil.IsNumeric(id) && !slices.Contains(ids, id) {
			ids = append(ids, id)
		}
	})
	return ids, nil
}
