package requtil

// Redirect makes Wrap answer 302 Found to URL instead of a JSON body.
type Redirect struct {
	URL string
}
