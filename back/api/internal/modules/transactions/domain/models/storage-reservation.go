package models

import "time"

// StorageReservation is the disk space held for the tree of an upload root until it finishes.
type StorageReservation struct {
	RootId    string
	UserId    int
	Bytes     int64
	CreatedAt time.Time
}
