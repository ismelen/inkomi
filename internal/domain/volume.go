package domain

import (
	"sync"
	"time"
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

func (v *Volume) GetConversionDuration() time.Duration {
	var size int64
	for _, chap := range v.Chapters {
		size += chap.Size
	}
	size = size >> 20
	size = int64(float64(size)*0.9)
	
	return time.Duration(size*int64(time.Second))
}