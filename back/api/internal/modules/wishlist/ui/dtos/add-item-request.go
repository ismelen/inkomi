package dtos

type AddItemRequest struct {
	Kind      string `json:"kind"`
	Title     string `json:"title"`
	Author    string `json:"author"`
	WorkKey   string `json:"workKey"`
	AuthorKey string `json:"authorKey"`
	// CoverURL is the cover of the picked suggestion; only catalog covers are kept.
	CoverURL string `json:"coverUrl"`
}
