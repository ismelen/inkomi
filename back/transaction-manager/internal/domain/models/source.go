package models

import "time"

type SourceStatus string

const (
	SourceStatusPendingUpload SourceStatus = "pending_upload"
	SourceStatusQueued        SourceStatus = "queued"
	SourceStatusProcessing    SourceStatus = "processing"
	SourceStatusJoining       SourceStatus = "joining"
	SourceStatusKepubifying   SourceStatus = "kepubifying"
	SourceStatusSent          SourceStatus = "sent"
	SourceStatusFailed        SourceStatus = "failed"
)

type ReadingDirection string

const (
	ReadingDirectionLTR ReadingDirection = "ltr"
	ReadingDirectionRTL ReadingDirection = "rtl"
)

type SourceType string

const (
	SourceTypeFolder  SourceStatus = "folder"
	SourceTypeFile    SourceStatus = "file"
	SourceTypeLibrary SourceStatus = "library"
)

type Source struct {
	Id               string           `json:"id"`
	UserId           int              `json:"userId"`
	Size             *int64           `json:"size"`
	Filename         string           `json:"filename"`
	Title            string           `json:"title"`
	Type             SourceType       `json:"type"`
	Status           SourceStatus     `json:"status"`
	Error            *string          `json:"error"`
	CreatedAt        time.Time        `json:"createdAt"`
	UpdatedAt        time.Time        `json:"updatedAt"`
	CompletedAt      *time.Time       `json:"completedAt"`
	ShouldJoin       bool             `json:"shouldJoin"`
	ReadingDirection ReadingDirection `json:"readingDirection"`
	FolderId         *string          `json:"folderId"`
	ConfigHash       *string          `json:"configHash"`
	Kepubify         bool             `json:"kepubify"`
}
