package MangaController

import (
	"fmt"
	"io"
	"ismelen/ermc/internal/manga/domain/MangaModels"
	MangaDtos "ismelen/ermc/internal/manga/domain/dtos"
	MangaConverter "ismelen/ermc/internal/manga/logic/coverters/manga"
	SharedInterfaces "ismelen/ermc/internal/shared/logic/interfaces"
	SysUtils "ismelen/ermc/internal/utils/sys"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func convertManga(c echo.Context, statuses *sync.Map) error {
	dto := new(MangaDtos.MangaConvertRequestDTO)
	if err := c.Bind(dto); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error": "Datos invalidos",
		})
	}

	tempDir, err := storeFormFiles(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error": err.Error(),
		})
	}

	opts := MangaModels.NewOptions(tempDir, dto.Profile, dto.Title, dto.Author, dto.Merge)
	if err := opts.ValidateAndNormalize(); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"error": err.Error(),
		})
	}

	id := uuid.New().String()

	c.JSON(http.StatusOK, echo.Map{
		"id": id,
	})

	notifier := SharedInterfaces.Notifier{}
	observer := MangaConverter.NewObserver()
	converter := MangaConverter.New(&opts, &notifier)
	go func() {
		defer os.RemoveAll(tempDir)
		converter.Convert()
	}()

	updateStatus(id, statuses, observer)

	return nil
}

func updateStatus(id string, statuses *sync.Map, observer MangaConverter.MangaConverterObserver) {
	status := &statusData{Msg: "0", Err: "", Paths: []string{}}
	statuses.Store(id, status)
	go func() {
		paths := observer.Listen(func(text string, err error) {
			if err != nil {
				status.Err = err.Error()
				return
			}
			status.Msg = text
		})
		status.Paths = paths
	}()
}

func storeFormFiles(c echo.Context) (string, error) {
	form, err := c.MultipartForm()
	if err != nil {
		return "", err
	}
	files := form.File["files"]
	if len(files) == 0 {
		return "", fmt.Errorf("No files")
	}
	tempDir := SysUtils.NewTempDir("request")

	for _, file := range files {
		src, err := file.Open()
		if err != nil {
			return "", err
		}
		defer src.Close()

		dstPath := filepath.Join(tempDir, file.Filename)
		dst, err := os.Create(dstPath)
		if err != nil {
			return "", err
		}
		defer dst.Close()

		if _, err = io.Copy(dst, src); err != nil {
			return "", err
		}
	}

	return tempDir, nil
}
