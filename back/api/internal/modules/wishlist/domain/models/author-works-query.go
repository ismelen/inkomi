package models

// AuthorWorksQuery asks for a page of the works of an author, by its catalog key or, for
// authors typed by hand, by name. Language is an ISO 639-1 code; empty keeps every language.
type AuthorWorksQuery struct {
	AuthorKey string
	Author    string
	Sort      WorksSort
	Language  string
	Page      int
}
