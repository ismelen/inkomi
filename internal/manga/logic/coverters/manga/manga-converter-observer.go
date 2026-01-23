package MangaConverter

import (
	"fmt"
	"net/http"
	"time"

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

func (this MangaConverterObserver) OnNotify(event any) {
	e := event.(MangaConverterEvent)
	this.EventChan <- e
}


func (this *MangaConverterObserver) ListenAndShow() []string {
	return this.Listen(func(text string, err error) {
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		fmt.Println(text)
	})
}

func (this *MangaConverterObserver) ListenAndFlush(c echo.Context, timeSpan time.Duration) []string {
	ticker := time.NewTicker(timeSpan)
	defer ticker.Stop()
	paths := []string{}

	var lastMsg *string
	errChan := make(chan string)

	ctx := c.Request().Context()

	go func() {
		paths = this.Listen(func(text string, err error) {
			if err != nil {
				errChan <- err.Error()
				return
			}
			lastMsg = &text
		})
	}()

	for {
		if len(paths) > 0 { break }
		select{
			case <-ctx.Done():{
				return nil
			}
			case  <- ticker.C: {
				if lastMsg == nil { continue }
				fmt.Fprintf(c.Response(), "data: %s\n\n", *lastMsg)
				c.Response().Flush()
				lastMsg = nil
			}
			case err := <- errChan: {
				c.JSON(http.StatusInternalServerError, echo.Map{
					"error": err,
				})
			}
		}
	}

	return paths
}

func (this *MangaConverterObserver) Listen(onEvent func(text string, err error)) []string {
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
