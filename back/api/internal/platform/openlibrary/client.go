// Package openlibrary is a small client of the Open Library search APIs (books and authors),
// with a single rate limit for the whole process.
package openlibrary

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"golang.org/x/time/rate"
)

const (
	DefaultBaseURL = "https://openlibrary.org"
	maxBodySize    = 4 << 20
)

const (
	searchFields  = "key,title,author_name,author_key,subject,first_publish_year,language,cover_i"
	editionFields = ",editions,editions.title,editions.cover_i"
)

// marcLanguages maps the ISO 639-1 codes the clients use to the MARC codes of Open Library.
var marcLanguages = map[string]string{
	"es": "spa", "en": "eng", "fr": "fre", "de": "ger", "it": "ita", "pt": "por", "ca": "cat",
	"ru": "rus", "ja": "jpn", "zh": "chi", "nl": "dut", "pl": "pol", "tr": "tur",
}

// IsLanguage reports whether code (ISO 639-1) can filter a search.
func IsLanguage(code string) bool {
	_, ok := marcLanguages[code]
	return ok
}

// Client calls Open Library at most rps times per second. Open Library allows about one request
// per second to anonymous clients and three to the ones that identify themselves in the
// User-Agent with a contact.
type Client struct {
	baseURL   string
	userAgent string
	http      *http.Client
	limiter   *rate.Limiter
}

func NewClient(baseURL, userAgent string, rps float64) *Client {
	if rps <= 0 {
		rps = 1
	}
	return &Client{
		baseURL:   strings.TrimRight(baseURL, "/"),
		userAgent: userAgent,
		http:      &http.Client{Timeout: 20 * time.Second},
		limiter:   rate.NewLimiter(rate.Limit(rps), 1),
	}
}

// SearchQuery is a query of search.json; empty fields are left out.
type SearchQuery struct {
	Title     string
	Author    string
	AuthorKey string
	// Query is free text matched against titles and authors.
	Query string
	// Sort is a sort of search.json (new, old, title, readinglog...); empty sorts by relevance.
	Sort string
	// Language (ISO 639-1, see IsLanguage) keeps the works with an edition in it, and reads
	// the title and cover of that edition.
	Language string
	Limit    int
	// Page starts at 1.
	Page int
}

// SearchResult is a page of works and how many works match in total.
type SearchResult struct {
	Works []Work
	Total int
}

func (c *Client) SearchWorks(ctx context.Context, q SearchQuery) (SearchResult, error) {
	params := url.Values{}
	set := func(key, value string) {
		if value != "" {
			params.Set(key, value)
		}
	}
	set("title", q.Title)
	set("author", q.Author)
	set("author_key", q.AuthorKey)
	set("q", q.Query)
	set("sort", q.Sort)
	fields := searchFields
	if marc, ok := marcLanguages[q.Language]; ok {
		params.Set("language", marc)
		params.Set("lang", q.Language)
		fields += editionFields
	}
	params.Set("fields", fields)
	params.Set("limit", strconv.Itoa(max(q.Limit, 1)))
	if q.Page > 1 {
		params.Set("page", strconv.Itoa(q.Page))
	}

	var body struct {
		NumFound int         `json:"numFound"`
		Docs     []searchDoc `json:"docs"`
	}
	if err := c.getJSON(ctx, "/search.json?"+params.Encode(), &body); err != nil {
		return SearchResult{}, err
	}
	works := make([]Work, 0, len(body.Docs))
	for _, doc := range body.Docs {
		if work, ok := doc.work(); ok {
			works = append(works, work)
		}
	}
	return SearchResult{Works: works, Total: body.NumFound}, nil
}

func (c *Client) SearchAuthors(ctx context.Context, query string, limit int) ([]Author, error) {
	params := url.Values{"q": {query}, "limit": {strconv.Itoa(max(limit, 1))}}
	var body struct {
		Docs []struct {
			Key       string `json:"key"`
			Name      string `json:"name"`
			TopWork   string `json:"top_work"`
			WorkCount int    `json:"work_count"`
		} `json:"docs"`
	}
	if err := c.getJSON(ctx, "/search/authors.json?"+params.Encode(), &body); err != nil {
		return nil, err
	}
	authors := make([]Author, 0, len(body.Docs))
	for _, doc := range body.Docs {
		if doc.Key == "" || doc.Name == "" {
			continue
		}
		authors = append(authors, Author{Key: trimKey(doc.Key), Name: doc.Name, TopWork: doc.TopWork, WorkCount: doc.WorkCount})
	}
	return authors, nil
}

// CoverURL is the medium cover image of a cover id, or "" without one.
func CoverURL(coverId int) string {
	if coverId <= 0 {
		return ""
	}
	return "https://covers.openlibrary.org/b/id/" + strconv.Itoa(coverId) + "-M.jpg"
}

func (c *Client) getJSON(ctx context.Context, path string, out any) error {
	if err := c.limiter.Wait(ctx); err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return fmt.Errorf("build request %s: %w", path, err)
	}
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("GET %s: %w", path, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("GET %s: %w", path, ErrNotFound)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GET %s: HTTP %d", path, resp.StatusCode)
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, maxBodySize)).Decode(out); err != nil {
		return fmt.Errorf("decode %s: %w", path, err)
	}
	return nil
}
