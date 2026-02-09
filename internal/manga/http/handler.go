package manga

import (
	"ismelen/ermc/internal/cloud"
	"ismelen/ermc/internal/domain"
	"ismelen/ermc/internal/manga"
	"ismelen/ermc/internal/pkg"
	volumeBuilder "ismelen/ermc/internal/volume-builder"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
)

type Handler struct{}

func NewHandler(serv *echo.Echo) *Handler {
	handler := &Handler{}

	serv.POST("/manga/convert", handler.handleConvert)
	serv.GET("/manga/:dir/:filename", handler.download)

	return handler
}

func (h *Handler) download(c echo.Context) error {
	dir := c.Param("dir")
	filename := c.Param("filename")

	path := filepath.Join(os.TempDir(), dir, filename)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return c.JSON(http.StatusNotFound, echo.Map{"message": "Archivo no encontrado"})
	}

	err := c.Attachment(path, filename)
	if err == nil {
		_ = os.Remove(path)
	}

	dirPath := filepath.Join(os.TempDir(), dir)
	if pkg.IsDirEmpty(dirPath) {
		os.RemoveAll(dirPath)
	}

	return nil
}

func (h *Handler) handleConvert(c echo.Context) error {
	//TODO: Get estimated time per each volume or file, return file link path (optional) and its aprox finish time
	
	dto := new(ConverterRequestDTO)
	if err := c.Bind(dto); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	
	cloudType := "google"
	if dto.CloudToken == "" {
		cloudType = "local"
	}

	cloudService, err := cloud.GetCloud(cloudType)
	if err != nil {
		return err
	}
	cloudService.Init(dto.CloudToken, dto.CloudFolder)

	settings, err := domain.NewSettings(
		dto.Author,
		dto.Title,
		dto.Profile,
		dto.Merge,
		dto.FirstVolumeNum,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	volumes, err := getVolumes(c, settings)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	type PathTime struct {
		Path string `json:"path"`
		EndTime time.Time `json:"endTime"`
	}
	
	var paths []PathTime
	endTime := time.Now()
	for _, volume := range volumes {
		var path string
		if (cloudType == "local") {
			path = filepath.Join(settings.Output.Base, volume.Name)
			if(settings.Profile.IsKepub){
				path += ".kepub"
			}
			path += ".epub"
		} else {
			path = volume.Name
		}

		endTime = endTime.Add(volume.GetConversionDuration())

		paths = append(paths, PathTime{
			Path: path,
			EndTime: endTime,
		})
	}

	go func() {
		start := time.Now()
		
		settings.SetImageSettings(domain.NewDefaultImageSettings())
		settings.SetVolumes(volumes)

		ramLimit, err := strconv.Atoi(os.Getenv("RAM"))
		if err != nil {
			c.Echo().Logger.Error(err)
			ramLimit = 0
		}
		
		resultChan := make(chan string)	
		converter := manga.NewConverter(settings, int64(ramLimit), resultChan)
		go converter.Convert(dto.Format)
		
		for path := range resultChan {
			_, err = cloudService.Upload(path)
			if err != nil {
				c.Echo().Logger.Error(err)
			}
		}

		c.Echo().Logger.Print(time.Since(start).String())
	}()

	return c.JSON(http.StatusOK, echo.Map{
		"paths": paths,
	})
}

func getVolumes(c echo.Context, settings *domain.Settings) ([]*domain.Volume, error) {
	form, err := c.MultipartForm()
	if err != nil {
		return nil, err
	}

	files := form.File["files"]
	if len(files) == 0 {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "No files attached")
	}

	slices.SortFunc(files, func (a, b *multipart.FileHeader) int {
		return pkg.FilenameCmp(a.Filename, b.Filename)
	})

	filesExt := filepath.Ext(files[0].Filename)
	volumeBuilder, err := volumeBuilder.GetBuilder(filesExt)
	if err != nil {
		return nil, err
	}

	return volumeBuilder.FromMultipart(settings, files...)
}
