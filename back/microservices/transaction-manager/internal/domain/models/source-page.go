package models

import "time"

type SourcePage struct {
	Trees []SourceTree
	// NextCursor is nil when there are no older roots.
	NextCursor *SourceCursor
	SyncedAt   time.Time
}
