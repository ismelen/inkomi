package domain

type TransactionResponseDTO struct {
	Id       string `json:"id"`
	Title    string `json:"title"`
	Filename string `json:"filename"`
}

func NewTransactionResponseDTO(id, title, filename string) TransactionResponseDTO {
	return TransactionResponseDTO{Id: id, Title: title, Filename: filename}
}
