package MangaConverter

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

type MangaConverterObserver struct {
	EventChan chan MangaConverterEvent
}

func NewObserver() MangaConverterObserver {
	return MangaConverterObserver{
		EventChan: make(chan MangaConverterEvent),
	}
}

func (this *MangaConverterObserver) OnNotify(event any) {
	e := event.(MangaConverterEvent)
	this.EventChan <- e
}


func (this *MangaConverterObserver) ListenAndShow() []string {
	return this.listen(func(text string, err error) {
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		fmt.Println(text)
	})
}

func (this *MangaConverterObserver) ListenAndFlush(w http.ResponseWriter, c echo.Context) []string {
	return this.listen(func(text string, err error) {
		if err != nil {
			c.JSON(http.StatusInternalServerError, echo.Map{
				"error": err.Error(),
			})
			return
		}
		
		fmt.Fprintf(w, "data: %s\n\n", text)
		w.(http.Flusher).Flush()
	})
}

func (this *MangaConverterObserver) listen(onEvent func(text string, err error)) []string {
	var paths []string
	var cantChapters, totalChapters, pagesCompleted int
	var avgPagesPerChapter float64


	for event := range this.EventChan {
		switch event.Type {
		case EventStart: 
			totalChapters = event.Cant

		case EventChapterStart:
			cantChapters++
			avgPagesPerChapter += ((float64(event.Cant) - avgPagesPerChapter)/float64(cantChapters))

		case EventPageFinished:
			pagesCompleted++;
			if cantChapters != 0 {
				progress := (float64(pagesCompleted) * 100.0) / (float64(totalChapters) * float64(avgPagesPerChapter))
				onEvent(fmt.Sprintf("%.2f", progress), nil)
			}

		case EventError:
			onEvent("", event.Err)
			return nil

		case EventDone:
			onEvent(fmt.Sprintf("%.2f", 100.0), nil)
			close(this.EventChan)
			paths = event.Paths
		}
	}

	return paths
}
