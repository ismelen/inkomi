package models

// WorkVector is what ranking needs of an embedded work.
type WorkVector struct {
	Key              string
	Title            string
	AuthorKeys       []string
	Languages        []string
	FirstPublishYear int
	Embedding        Embedding
}
