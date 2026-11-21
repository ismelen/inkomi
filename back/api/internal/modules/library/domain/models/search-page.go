package models

// SearchPage is a page of search results. NextPage is the page to ask for to go on, or 0 when
// there are no more results.
type SearchPage struct {
	Books    []Book
	NextPage int
}
