package models

// SearchField is the book field a search query is matched against.
type SearchField string

const (
	SearchFieldTitle  SearchField = "title"
	SearchFieldAuthor SearchField = "author"
)

func (f SearchField) IsValid() bool {
	return f == SearchFieldTitle || f == SearchFieldAuthor
}
