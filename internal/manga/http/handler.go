package manga

import (
	"ismelen/ermc/internal/domain"
	"ismelen/ermc/internal/manga"
	volumeBuilder "ismelen/ermc/internal/volume-builder"
	"net/http"
	"path/filepath"

	"github.com/labstack/echo/v4"
)

type Handler struct{}

func NewHandler(serv *echo.Echo) *Handler {
	handler := &Handler{}

	serv.POST("/manga/convert", handler.handleConvert)

	return handler
}

func (h *Handler) handleConvert(c echo.Context) error {
	dto := new(ConverterRequestDTO)
	if err := c.Bind(dto); err != nil {
		return err
	}

	settings, err := domain.NewSettings(
		dto.Author,
		dto.Title,
		dto.Profile,
		dto.Merge,
		dto.FirstVolumeNum,
	)
	if err != nil {
		return err
	}

	volumes, err := getVolumes(c, settings)
	if err != nil {
		return err
	}

	settings.SetImageSettings(domain.NewDefaultImageSettings())
	settings.SetVolumes(volumes)

	converter := manga.NewConverter(settings, 1000)
	paths, err := converter.Convert(dto.Format)
	if err != nil {
		return err
	}

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

	filesExt := filepath.Ext(files[0].Filename)
	volumeBuilder, err := volumeBuilder.GetBuilder(filesExt)
	if err != nil {
		return nil, err
	}

	return volumeBuilder.FromMultipart(settings, files...)
}
