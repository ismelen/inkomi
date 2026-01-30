package domain

import (
	"sync"
)

type Volume struct {
	Name     string
	Chapters []*Chapter
	Wg       *sync.WaitGroup
}

func NewVolume(name string, chapters ...*Chapter) *Volume {
	return &Volume{
		Name:     name,
		Chapters: chapters,
		Wg: &sync.WaitGroup{},
	}
}
