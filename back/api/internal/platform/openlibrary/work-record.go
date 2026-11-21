package openlibrary

import (
	"context"
	"encoding/json"
	"net/url"
)

// WorkRecord is the record of a work (/works/<key>.json). Description is the raw text, which
// may carry markdown links.
type WorkRecord struct {
	Key              string
	Title            string
	Description      string
	CoverIds         []int
	Subjects         []string
	FirstPublishDate string
}

// textValue is a field Open Library writes either as a string or as {"type", "value"}.
type textValue string

func (t *textValue) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		*t = textValue(s)
		return nil
	}
	var typed struct {
		Value string `json:"value"`
	}
	if err := json.Unmarshal(data, &typed); err != nil {
		return err
	}
	*t = textValue(typed.Value)
	return nil
}

// GetWork returns ErrNotFound when there is no work with key.
func (c *Client) GetWork(ctx context.Context, key string) (WorkRecord, error) {
	var body struct {
		Key              string    `json:"key"`
		Title            string    `json:"title"`
		Description      textValue `json:"description"`
		Covers           []int     `json:"covers"`
		Subjects         []string  `json:"subjects"`
		FirstPublishDate string    `json:"first_publish_date"`
	}
	if err := c.getJSON(ctx, "/works/"+url.PathEscape(key)+".json", &body); err != nil {
		return WorkRecord{}, err
	}
	covers := make([]int, 0, len(body.Covers))
	for _, id := range body.Covers {
		if id > 0 {
			covers = append(covers, id)
		}
	}
	return WorkRecord{
		Key:              trimKey(body.Key),
		Title:            body.Title,
		Description:      string(body.Description),
		CoverIds:         covers,
		Subjects:         body.Subjects,
		FirstPublishDate: body.FirstPublishDate,
	}, nil
}
