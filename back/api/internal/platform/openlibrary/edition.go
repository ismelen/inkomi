package openlibrary

import (
	"context"
	"net/url"
)

// EditionKeyIn returns the key (OL…M) of the best edition of the work in lang (ISO 639-1, see
// IsLanguage), or "" when the work has none.
func (c *Client) EditionKeyIn(ctx context.Context, workKey, lang string) (string, error) {
	marc, known := marcLanguages[lang]
	if !known {
		return "", nil
	}
	params := url.Values{
		"q":        {"key:/works/" + workKey},
		"language": {marc},
		"lang":     {lang},
		"fields":   {"key,editions,editions.key,editions.language"},
		"limit":    {"1"},
	}
	var body struct {
		Docs []struct {
			Editions struct {
				Docs []struct {
					Key       string   `json:"key"`
					Languages []string `json:"language"`
				} `json:"docs"`
			} `json:"editions"`
		} `json:"docs"`
	}
	if err := c.getJSON(ctx, "/search.json?"+params.Encode(), &body); err != nil {
		return "", err
	}
	for _, doc := range body.Docs {
		for _, edition := range doc.Editions.Docs {
			for _, language := range edition.Languages {
				if language == marc && edition.Key != "" {
					return trimKey(edition.Key), nil
				}
			}
		}
	}
	return "", nil
}
