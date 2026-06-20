package domain

import "time"

type Transaction struct {
	Id      string
	StartAt time.Time
	Path    string
	ResultPath string
	Size    int64
	Current int64
	Error error
	Done bool
	Canceled bool
}