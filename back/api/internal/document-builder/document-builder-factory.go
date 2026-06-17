package documentBuilder

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

func GetBuilder(format string) (BuilderI, error) {
	switch format {
	case "epub":
		return &EpubBuilder{}, nil
	default:
		return nil, echo.NewHTTPError(
			http.StatusInternalServerError,
			fmt.Sprintf("Format output not supported: %s", format),
		)
	}
}
