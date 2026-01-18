package MangaConverter

type EventType int

const (
	EventError EventType = iota
	EventDone
	EventChapterStart
	EventPageFinished
)

type MangaConverterEvent struct {
	Type  EventType
	Err   error
	Paths []string
	Cant  int
}
