package domain

type TransactionResponseDTO struct {
	Id    string `json:"id"`
	Title string `json:"title"`
}

func NewTransactionResponseDTO(id, title string) TransactionResponseDTO {
	return TransactionResponseDTO{Id: id, Title: title}
}
