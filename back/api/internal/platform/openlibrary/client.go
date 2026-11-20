// Package openlibrary is a small client of the Open Library APIs (search, authors, works and
// subjects) shared by the modules, with a single rate limit for the whole process.
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

const searchFields = "key,title,author_name,author_key,subject,first_publish_year,language,cover_i"

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
	Subject   string
	// Query is a free query, e.g. "first_publish_year:[2025 TO *]".
	Query string
	// Sort is empty (relevance) or "new".
	Sort  string
	Limit int
}

func (c *Client) SearchWorks(ctx context.Context, q SearchQuery) ([]Work, error) {
	params := url.Values{}
	set := func(key, value string) {
		if value != "" {
			params.Set(key, value)
		}
	}
	set("title", q.Title)
	set("author", q.Author)
	set("author_key", q.AuthorKey)
	set("subject", q.Subject)
	set("q", q.Query)
	set("sort", q.Sort)
	params.Set("fields", searchFields)
	params.Set("limit", strconv.Itoa(max(q.Limit, 1)))

	var body struct {
		Docs []searchDoc `json:"docs"`
	}
	if err := c.getJSON(ctx, "/search.json?"+params.Encode(), &body); err != nil {
		return nil, err
	}
	works := make([]Work, 0, len(body.Docs))
	for _, doc := range body.Docs {
		if work, ok := doc.work(); ok {
			works = append(works, work)
		}
	}
	return works, nil
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

// WorkDetail reads the description and subjects of a work (/works/<key>.json).
func (c *Client) WorkDetail(ctx context.Context, key string) (*Detail, error) {
	var body struct {
		Description json.RawMessage `json:"description"`
		Subjects    []string        `json:"subjects"`
	}
	if err := c.getJSON(ctx, "/works/"+url.PathEscape(trimKey(key))+".json", &body); err != nil {
		return nil, err
	}
	return &Detail{Description: description(body.Description), Subjects: body.Subjects}, nil
}

// SubjectWorks lists works of a subject (/subjects/<subject>.json).
func (c *Client) SubjectWorks(ctx context.Context, subject string, limit int) ([]Work, error) {
	slug := strings.ReplaceAll(strings.ToLower(strings.TrimSpace(subject)), " ", "_")
	var body struct {
		Works []struct {
			Key              string   `json:"key"`
			Title            string   `json:"title"`
			CoverId          int      `json:"cover_id"`
			FirstPublishYear int      `json:"first_publish_year"`
			Subject          []string `json:"subject"`
			Authors          []struct {
				Key  string `json:"key"`
				Name string `json:"name"`
			} `json:"authors"`
		} `json:"works"`
	}
	path := "/subjects/" + url.PathEscape(slug) + ".json?limit=" + strconv.Itoa(max(limit, 1))
	if err := c.getJSON(ctx, path, &body); err != nil {
		return nil, err
	}
	works := make([]Work, 0, len(body.Works))
	for _, w := range body.Works {
		if w.Key == "" || w.Title == "" {
			continue
		}
		work := Work{Key: trimKey(w.Key), Title: w.Title, CoverId: w.CoverId, FirstPublishYear: w.FirstPublishYear, Subjects: w.Subject}
		for _, a := range w.Authors {
			work.Authors = append(work.Authors, a.Name)
			work.AuthorKeys = append(work.AuthorKeys, trimKey(a.Key))
		}
		works = append(works, work)
	}
	return works, nil
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
