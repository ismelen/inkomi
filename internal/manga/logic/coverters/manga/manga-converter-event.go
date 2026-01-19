package MangaConverter

type EventType int

const (
	EventError EventType = iota
	EventDone
	EventChapterStart
	EventPageFinished
	EventStart
)

type MangaConverterEvent struct {
	Type  EventType
	Err   error
	Paths []string
	Cant  int
}
