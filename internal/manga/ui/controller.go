package MangaController

import (
	"io"
	MangaService "ismelen/ermc/internal/manga/logic"
	MangaDtos "ismelen/ermc/internal/manga/logic/dtos"
	manga "ismelen/ermc/internal/manga/logic/models"
	SysUtils "ismelen/ermc/internal/utils/sys"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/labstack/echo/v4"
)

func New(s *echo.Echo) {
	s.POST("/manga/convert", convert)
	s.GET("/manga/:id/:filename", downloadFile)
}

var jobs sync.Map

func convert(c echo.Context) error {
	dto := new(MangaDtos.MangaConvertRequestDTO)
	if err := c.Bind(dto); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Datos invalidos",
		})
	}

	form, err := c.MultipartForm()
	if err != nil {
		return err
	}
	files := form.File["files"]
	tempDir := SysUtils.NewTempDir("request")

	for _, file := range files {
		src, err := file.Open()
		if err != nil {
			return err
		}
		defer src.Close()

		dstPath := filepath.Join(tempDir, file.Filename)
		dst, err := os.Create(dstPath)
		if err != nil {
			return err
		}
		defer dst.Close()

		if _, err = io.Copy(dst, src); err != nil {
			return err
		}
	}
	
	opts := manga.NewOptions(tempDir, dto.Profile, dto.Title, dto.Author, dto.Merge)
	if err := opts.ValidateAndNormalize(); err != nil {
		return err
	}
	os.RemoveAll(tempDir)

	go MangaService.ProcessInputs(&opts)

	// w := c.Response().Writer
	// w.Header().Set("Content-Type", "text/event-stream")
	// w.Header().Set("Cache-Control", "no-cache")
	// w.Header().Set("Connection", "keep-alive")

	// for path := range results {
	// 	fmt.Fprintf(w, "%s", path)
	// 	w.(http.Flusher).Flush()
	// }

	// fmt.Fprintf(w, "%s", "done")
	// w.(http.Flusher).Flush()

	return nil
}
/*
{

}

*/

func downloadFile(c echo.Context) error {
	return nil
}


