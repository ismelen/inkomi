package manga

import (
	volumeBuilder "ismelen/ermc/v2/volume-builder"
	"net/http"
	"path/filepath"

	"github.com/labstack/echo/v4"
)

type Handler struct{}

func NewHandler(serv *echo.Echo) *Handler {
	handler := &Handler{}

	serv.POST("/manga/convert", handler.convert)

	return handler
}

func (h *Handler) convert(c echo.Context) error {
	dto := new(ConverterRequestDTO)
	if err := c.Bind(dto); err != nil {
		return err
	}

	form, err := c.MultipartForm()
	if err != nil {
		return err
	}

	files := form.File["files"]
	if len(files) == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "No files attached")
	}
	filesExt := filepath.Ext(files[0].Filename)

	volumeBuilder, err := volumeBuilder.GetBuilder(filesExt)
	if err != nil {
		return err
	}

	_, err = volumeBuilder.FromMultipart(files...)
	if err != nil {
		return err
	}

	return nil
}
