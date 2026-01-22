package MangaController

import (
	"io"
	MangaConverter "ismelen/ermc/internal/manga/logic/coverters/manga"
	MangaDtos "ismelen/ermc/internal/manga/logic/dtos"
	manga "ismelen/ermc/internal/manga/logic/models"
	SharedInterfaces "ismelen/ermc/internal/shared/logic/interfaces"
	SysUtils "ismelen/ermc/internal/utils/sys"
	"net/http"
	"os"
	"path/filepath"

	"github.com/labstack/echo/v4"
)

func New(s *echo.Echo) {
	s.POST("/manga/convert", convert)
	s.GET("/manga/:id/:filename", downloadFile)
}

func convert(c echo.Context) error {
	c.Logger().Info("Hola")
	dto := new(MangaDtos.MangaConvertRequestDTO)

	dto.Author = c.FormValue("author")
	// dto.GoogleCloudFolder = c.FormValue("googleCloudFolder")
	dto.Merge = c.FormValue("merge") == "true"
	dto.Profile = c.FormValue("profile")
	// dto.StartingVolumeCount, _ = strconv.Atoi(c.FormValue("startingVolumeCount"))
	dto.Title = c.FormValue("title")
	
	// if err := c.Bind(dto); err != nil {
	// 	return c.JSON(http.StatusBadRequest, echo.Map{
	// 		"error": "Datos invalidos",
	// 	})
	// }

	form, err := c.MultipartForm()
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error": err.Error(),
		})
	}
	files := form.File["files"]
	if len(files) == 0 {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"error": "No files",
		})
	}
	tempDir := SysUtils.NewTempDir("request")


	for _, file := range files {
		src, err := file.Open()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{
				"error": err.Error(),
			})
		}
		defer src.Close()

		dstPath := filepath.Join(tempDir, file.Filename)
		dst, err := os.Create(dstPath)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{
				"error": err.Error(),
			})
		}
		defer dst.Close()

		if _, err = io.Copy(dst, src); err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{
				"error": err.Error(),
			})
		}
	}

	opts := manga.NewOptions(tempDir, dto.Profile, dto.Title, dto.Author, dto.Merge)
	if err := opts.ValidateAndNormalize(); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"error": err.Error(),
		})
	}
	defer os.RemoveAll(tempDir)
	

	c.Response().Header().Set(echo.HeaderContentType, "text/event-stream")
	c.Response().Header().Set("Cache-Control", "no-cache")
	c.Response().Header().Set("Connection", "keep-alive")
	c.Response().WriteHeader(http.StatusOK)

	notifier := SharedInterfaces.Notifier{}
	observer := MangaConverter.NewObserver()
	notifier.Register(&observer)

	converter := MangaConverter.New(&opts, &notifier)
	go converter.Convert()

	paths := observer.ListenAndFlush(c)

	return c.JSON(http.StatusOK, echo.Map{
		"urls": paths,
	})
}

func downloadFile(c echo.Context) error {
	return nil
}


