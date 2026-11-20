package jobs

import "encoding/json"

// Msg is the payload published on every job Subject.
type Msg struct {
	Id       string `json:"id"`
	UserId   int    `json:"userId"`
	Filename string `json:"filename"`
	Title    string `json:"title"`
	Kepubify bool   `json:"kepubify"`
	ToCloud  bool   `json:"toCloud"`
	ErrorMsg string `json:"errorMsg,omitempty"`
	Size     int64  `json:"size,omitempty"`

	*Manga
	Parent *Parent `json:"parent,omitempty"`
	Result *Result `json:"result,omitempty"`
}

type Manga struct {
	Rtl      bool            `json:"rtl"`
	MangaCfg json.RawMessage `json:"mangaCfg,omitempty"`
}

type Parent struct {
	Id    string `json:"parentId"`
	Title string `json:"title"`
	Rtl   bool   `json:"rtl"`
	Total int    `json:"total"`
	Index int    `json:"index"`
}

// Result marks a Msg as one of the epubs produced by joining a parent folder.
type Result struct {
	ParentId string `json:"parentId"`
	Index    int    `json:"index"`
	Total    int    `json:"total"`
}
