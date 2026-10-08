package models

type ReadingDirection string

const (
	ReadingDirectionLTR = "ltr"
	ReadingDirectionRTL = "rtl"
)

type JoinMsg struct {
	Id        string           `json:"id"`
	UserId    int              `json:"userId"`
	Kepubify  bool             `json:"kepubify"`
	ToCloud   bool             `json:"toCloud"`
	Title     string           `json:"title"`
	Items     []JoinItem       `json:"items"`
	Direction ReadingDirection `json:"direction"`
}

type JoinItem struct {
	Id       string `json:"id"`
	Filename string `json:"filename"`
}
