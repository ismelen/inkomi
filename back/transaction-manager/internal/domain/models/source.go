package models

import (
	"path/filepath"
	"strings"
	"time"
)

type SourceStatus string

const (
	SourceStatusPendingUpload SourceStatus = "pending_upload"
	SourceStatusQueued        SourceStatus = "queued"
	SourceStatusProcessing    SourceStatus = "processing"
	SourceStatusWaitingJoin   SourceStatus = "waiting_join"
	SourceStatusJoining       SourceStatus = "joining"
	SourceStatusKepubifying   SourceStatus = "kepubifying"
	SourceStatusSent          SourceStatus = "sent"
	SourceStatusFailed        SourceStatus = "failed"
	SourceStatusDone          SourceStatus = "done"
)

// ReadingDirection and SourceType are part of the public API contract; they are
// persisted as source.isRtl and source.isFile.
type ReadingDirection string

const (
	ReadingDirectionLTR ReadingDirection = "ltr"
	ReadingDirectionRTL ReadingDirection = "rtl"
)

type SourceType string

const (
	SourceTypeFolder  SourceType = "folder"
	SourceTypeFile    SourceType = "file"
	SourceTypeLibrary SourceType = "library"
)

// Source maps the source table.
type Source struct {
	ID          string       `json:"id"`
	UserID      int          `json:"userId"`
	Title       string       `json:"title"`
	Filename    string       `json:"filename"`
	Kepubify    bool         `json:"kepubify"`
	ToCloud     bool         `json:"toCloud"`
	ConfigKey   *string      `json:"configKey"`
	IsRTL       bool         `json:"isRtl"`
	Join        bool         `json:"join"`
	Status      SourceStatus `json:"status"`
	Size        *int64       `json:"size"`
	IsFile      bool         `json:"isFile"`
	CreatedAt   time.Time    `json:"createdAt"`
	UpdatedAt   time.Time    `json:"updatedAt"`
	CompletedAt *time.Time   `json:"completedAt"`
	ParentID    *string      `json:"parentId"`
}

// IsLibrary reports whether the source is a library download: a file that is
// not uploaded by the client, so its size is unknown.
func (s *Source) IsLibrary() bool {
	return s.IsFile && (s.Size == nil || *s.Size == 0)
}

func (s *Source) IsEpub() bool {
	return strings.EqualFold(filepath.Ext(s.Filename), ".epub")
}

func (s *Source) IsKepub() bool {
	return strings.Contains(strings.ToLower(s.Filename), ".kepub")
}
