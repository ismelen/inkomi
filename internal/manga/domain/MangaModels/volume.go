package MangaModels

import "sync"

type Volume struct {
	Filename string
	Chapters []*Chapter
	Wg       *sync.WaitGroup
}

func NewVolume(filename string, chapters []*Chapter) Volume {
	return Volume{
		Filename: filename,
		Chapters: chapters,
		Wg:       &sync.WaitGroup{},
	}
}
