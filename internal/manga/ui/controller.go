package MangaController

import (
	"sync"

	"github.com/labstack/echo/v4"
)

type statusData struct {
	Msg string
	Err string
	Paths []string
}

func New(s *echo.Echo) {
	var statuses sync.Map
	s.GET("/manga/status/:id", func(c echo.Context) error {
		return conversionStatus(c, &statuses)
	})
	s.POST("/manga/convert", func(c echo.Context) error {
		return convertManga(c, &statuses)
	})
	s.GET("/manga/:id/:filename", downloadFile)
}

func downloadFile(c echo.Context) error {
	return nil
}


