// Package wikimedia finds the Wikipedia article of a book through its Open Library keys on
// Wikidata and reads the article summary. Neither API needs a key, but both ask for a
// User-Agent that identifies the client.
package wikimedia

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	DefaultSPARQLURL = "https://query.wikidata.org/sparql"
	// DefaultWikipediaURL is the base URL of each Wikipedia; {lang} is the language code.
	DefaultWikipediaURL = "https://{lang}.wikipedia.org"
	maxBodySize         = 1 << 20
)

var errNotFound = errors.New("not found")

type Client struct {
	sparqlURL    string
	wikipediaURL string
	userAgent    string
	http         *http.Client
}

func NewClient(sparqlURL, wikipediaURL, userAgent string) *Client {
	return &Client{
		sparqlURL:    sparqlURL,
		wikipediaURL: strings.TrimRight(wikipediaURL, "/"),
		userAgent:    userAgent,
		http:         &http.Client{Timeout: 15 * time.Second},
	}
}

// ArticleFor returns the title of the Wikipedia article in lang (ISO 639-1) about the book with
// any of the Open Library keys (work OL…W or edition OL…M; an edition counts through the work it
// is an edition of), or "" when there is none.
func (c *Client) ArticleFor(ctx context.Context, openLibraryKeys []string, lang string) (string, error) {
	values := make([]string, 0, len(openLibraryKeys))
	for _, key := range openLibraryKeys {
		if key != "" {
			values = append(values, `"`+strings.ReplaceAll(key, `"`, "")+`"`)
		}
	}
	if len(values) == 0 {
		return "", nil
	}
	site := "https://" + lang + ".wikipedia.org/"
	query := `SELECT ?article WHERE {
		VALUES ?id { ` + strings.Join(values, " ") + ` }
		?x wdt:P648 ?id ; wdt:P629? ?work .
		?article schema:about ?work ; schema:isPartOf <` + site + `> .
	} LIMIT 1`

	var body struct {
		Results struct {
			Bindings []struct {
				Article struct {
					Value string `json:"value"`
				} `json:"article"`
			} `json:"bindings"`
		} `json:"results"`
	}
	endpoint := c.sparqlURL + "?" + url.Values{"query": {query}, "format": {"json"}}.Encode()
	if err := c.getJSON(ctx, endpoint, "application/sparql-results+json", &body); err != nil {
		return "", err
	}
	for _, binding := range body.Results.Bindings {
		article := binding.Article.Value
		if i := strings.Index(article, "/wiki/"); i >= 0 {
			title, err := url.PathUnescape(article[i+len("/wiki/"):])
			if err == nil && title != "" {
				return title, nil
			}
		}
	}
	return "", nil
}

// Summary returns the plain-text lead of the article, or "" when it is missing or is a
// disambiguation page.
func (c *Client) Summary(ctx context.Context, lang, title string) (string, error) {
	base := strings.ReplaceAll(c.wikipediaURL, "{lang}", lang)
	endpoint := base + "/api/rest_v1/page/summary/" + url.PathEscape(strings.ReplaceAll(title, " ", "_"))
	var body struct {
		Type    string `json:"type"`
		Extract string `json:"extract"`
	}
	err := c.getJSON(ctx, endpoint, "application/json", &body)
	if errors.Is(err, errNotFound) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	if body.Type != "standard" {
		return "", nil
	}
	return strings.TrimSpace(body.Extract), nil
}

func (c *Client) getJSON(ctx context.Context, endpoint, accept string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", accept)

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("GET %s: %w", req.URL.Host, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return errNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GET %s%s: HTTP %d", req.URL.Host, req.URL.Path, resp.StatusCode)
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, maxBodySize)).Decode(out); err != nil {
		return fmt.Errorf("decode %s: %w", req.URL.Host, err)
	}
	return nil
}
