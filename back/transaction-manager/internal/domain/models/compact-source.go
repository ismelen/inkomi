package models

type CompactSoruce struct {
	Id               string           `json:"id"`
	UserId           int              `json:"userId"`
	Filename         string           `json:"filename,omitempty,omitzero"`
	Type             SourceType       `json:"type,omitempty,omitzero"`
	Title            string           `json:"title,omitempty,omitzero"`
	ShouldJoin       bool             `json:"shouldJoin,omitempty,omitzero"`
	ReadingDirection ReadingDirection `json:"readingDirection,omitempty,omitzero"`
	FolderId         *string          `json:"folderId,omitempty,omitzero"`
	Config           *Config          `json:"config,omitempty,omitzero"`
	Kepubify         bool             `json:"kepubify,omitempty"`
}

type SourceItem struct {
	Id       string `json:"id"`
	Filename string `json:"filename"`
	Size     *int64 `json:"size"`
}

type FolderWithItems struct {
	Folder CompactSoruce
	Items  []SourceItem
}
