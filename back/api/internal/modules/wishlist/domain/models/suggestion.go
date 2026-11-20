package models

// Suggestion is a book or an author of the catalog that matches what the user is typing.
type Suggestion struct {
	Kind      ItemKind
	Title     string
	Author    string
	WorkKey   string
	AuthorKey string
	Year      int
	CoverURL  string
}
