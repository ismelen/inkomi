package domain

type Book struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Author      string `json:"author"`
	Year        string `json:"year"`
	Publisher   string `json:"publisher"`
	Pages       string `json:"pages"`
	Language    string `json:"language"`
	SizeStr     string `json:"filesize_str,omitempty"`
	Extension   string `json:"extension"`
	MD5         string `json:"md5"`
	CoverURL    string `json:"cover_url"`
	DownloadURL string `json:"download_url"`
}
