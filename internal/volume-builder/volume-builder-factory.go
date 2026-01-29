package volumeBuilder

import (
	"fmt"
	fromCbz "ismelen/ermc/internal/volume-builder/from-cbz"
	"net/http"

	"github.com/labstack/echo/v4"
)

func GetBuilder(fileExt string) (BuilderI, error) {
	switch fileExt {
	case "cbz":
		return &fromCbz.VolumeFromCbz{}, nil
	default:
		return nil, echo.NewHTTPError(
			http.StatusInternalServerError,
			fmt.Sprintf("File format not supported: %s", fileExt),
		)
	}
}
