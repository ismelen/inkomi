package models

import "encoding/json"

// Subjects published by the transaction manager to dispatch a source.
const (
	SubjectJobManga    = "job.init.manga"
	SubjectJobLibrary  = "job.init.library"
	SubjectJobWaitJoin = "job.step.waitjoin"
	SubjectJobKepub    = "job.step.kepub"
	SubjectJobDone     = "job.step.done"
)

// JobMsg is the payload sent to the workers. The base fields are always present;
// MangaJob is flattened in for the manga worker and Parent is set when the source
// belongs to a folder that has to be joined.
type JobMsg struct {
	ID       string `json:"id"`
	UserID   int    `json:"userId"`
	Filename string `json:"filename"`
	Title    string `json:"title"`
	Kepubify bool   `json:"kepubify"`
	ToCloud  bool   `json:"toCloud"`

	*MangaJob
	Parent *JobParent `json:"parent,omitempty"`
}

type MangaJob struct {
	RTL      bool            `json:"rtl"`
	MangaCfg json.RawMessage `json:"mangaCfg,omitempty"`
}

type JobParent struct {
	ID    string `json:"parentId"`
	Title string `json:"title"`
	RTL   bool   `json:"rtl"`
}
