package manga

import (
	"fmt"
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

	return handler
}

func (h *Handler) handleConvert(c echo.Context) error {
	start := time.Now()
	
	dto := new(ConverterRequestDTO)
	if err := c.Bind(dto); err != nil {
		fmt.Println(err)
		return err
	}

	
	cloudService := "google"
	if dto.CloudToken == "" {
		cloudService = "tempfile"
	}

	cloud, err := cloud.GetCloud(cloudService)
	if err != nil {
		// TODO: Return user
	}
	cloud.Init(dto.CloudToken, dto.CloudFolder)

	settings, err := domain.NewSettings(
		dto.Author,
		dto.Title,
		dto.Profile,
		dto.Merge,
		dto.FirstVolumeNum,
	)
	if err != nil {
		fmt.Println(err)
		return err
	}

	volumes, err := getVolumes(c, settings)
	if err != nil {
		fmt.Println(err)
		return err
	}

	settings.SetImageSettings(domain.NewDefaultImageSettings())
	settings.SetVolumes(volumes)

	ramLimit, err := strconv.Atoi(os.Getenv("RAM"))
	if err != nil {
		ramLimit = 0
	}
	converter := manga.NewConverter(settings, int64(ramLimit))
	paths, err := converter.Convert(dto.Format)
	if err != nil {
		fmt.Println(err)
		return err
	}

	for _, path := range paths {
		if err := cloud.Upload(path); err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{
				"error": "Couldn't upload files to cloud",
			})
		}
	}

	return c.JSON(http.StatusOK, echo.Map{
		"paths": paths,
		"elapsed": time.Since(start),
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
