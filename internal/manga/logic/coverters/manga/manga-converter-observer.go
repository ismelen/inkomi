package MangaConverter

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

type MangaConverterObserver struct {
	EventChan chan MangaConverterEvent
}

func (this *MangaConverterObserver) OnNotify(event any) {
	e := event.(MangaConverterEvent)
	this.EventChan <- e
}

func (this *MangaConverterObserver) ListenHttp(w http.ResponseWriter, c echo.Context) []string {
	var paths []string
	var totalPages, pagesCompleted int

	c.Logger().Info("Started logger")

	for event := range this.EventChan {
		c.Logger().Info("New event")
		switch event.Type {
		case EventChapterStart:
			totalPages++

		case EventPageFinished:
			pagesCompleted += event.Cant

		case EventError:
			c.JSON(http.StatusInternalServerError, echo.Map{"error": event.Err.Error()})
			return nil

		case EventDone:
			close(this.EventChan)
			paths = event.Paths
		}

		if totalPages != 0 {
			progress := float64(pagesCompleted) * 100 / float64(totalPages)
			// fmt.Printf("%.2f", progress)
			c.Logger().Infof("%.2f", progress)

			progressStr := fmt.Sprintf("%.2f", progress)
			fmt.Fprintf(w, "data: %s\n\n", progressStr)
			w.(http.Flusher).Flush()
		}
	}

	return paths
}
