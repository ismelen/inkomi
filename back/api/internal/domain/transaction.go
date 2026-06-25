package domain

import "time"

type Transaction struct {
	Id         string
	StartAt    time.Time
	Path       string
	ResultPath string
	Pages      int
	Current    int
	Error      error
	Done       bool
	Canceled   bool
}
