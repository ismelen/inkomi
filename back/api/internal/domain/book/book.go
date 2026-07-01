package book

type Book struct {
	Title       string `json:"title"`
	Author      string `json:"author"`
	Pages       int    `json:"pages"`
	Language    string `json:"language"`
	Extension   string `json:"extension"`
	MD5         string `json:"md5"`
	CoverURL    string `json:"cover_url"`
	DownloadURL string `json:"download_url"`
}
