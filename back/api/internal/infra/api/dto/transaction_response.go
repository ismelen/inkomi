package dto

// TransactionResponse is the HTTP JSON response for a started conversion.
type TransactionResponse struct {
	Id       string `json:"id"`
	Title    string `json:"title"`
	Filename string `json:"filename"`
}
