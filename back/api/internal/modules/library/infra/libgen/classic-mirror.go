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
	"github.com/ismelen/inkomi/back/api/internal/modules/library/domain/models"
	"github.com/ismelen/inkomi/back/api/internal/modules/library/shared/strutil"
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
	CoverURL  string `json:"coverurl"`
}

func (c *ClassicMirror) Search(ctx context.Context, query string, field models.SearchField, page int) (models.SearchPage, error) {
	page = max(page, 1)
	ids, err := c.searchIds(ctx, query, field, page)
	if err != nil {
		return models.SearchPage{}, err
	}
	if len(ids) == 0 {
		return models.SearchPage{Books: []models.Book{}}, nil
	}
	next := 0
	if len(ids) >= maxResults {
		next = page + 1
	}

	books, err := c.metadata(ctx, ids)
	if err != nil {
		return models.SearchPage{}, err
	}
	return models.SearchPage{Books: books, NextPage: next}, nil
}

func (c *ClassicMirror) metadata(ctx context.Context, ids []string) ([]models.Book, error) {

	params := url.Values{}
	params.Set("ids", strings.Join(ids, ","))
	params.Set("fields", "id,title,author,pages,language,extension,md5,coverurl")
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
			CoverURL:  c.coverURL(b.CoverURL),
		})
	}
	return books, nil
}

// coverURL turns the coverurl of json.php (a path under /covers) into an absolute url.
func (c *ClassicMirror) coverURL(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}
	cover, err := resolveURL(c.url+"/covers/", strings.TrimPrefix(path, "/"))
	if err != nil {
		return ""
	}
	return cover
}

func classicColumn(field models.SearchField) string {
	if field == models.SearchFieldAuthor {
		return "author"
	}
	return "title"
}

func (c *ClassicMirror) searchIds(ctx context.Context, query string, field models.SearchField, page int) ([]string, error) {
	params := url.Values{}
	params.Set("req", query)
	params.Set("res", strconv.Itoa(maxResults))
	params.Set("view", "simple")
	params.Set("phrase", "1")
	params.Set("column", classicColumn(field))
	params.Set("page", strconv.Itoa(page))

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
