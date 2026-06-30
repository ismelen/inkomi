package book

type Book struct {
	ID          string `json:"id,omitempty"`
	Title       string `json:"title"`
	Author      string `json:"author"`
	Pages       string `json:"pages"`
	Language    string `json:"language"`
	Extension   string `json:"extension"`
	MD5         string `json:"md5"`
	CoverURL    string `json:"cover_url"`
	DownloadURL string `json:"download_url"`
}
