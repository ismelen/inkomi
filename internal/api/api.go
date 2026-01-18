package api

import (
	MangaController "ismelen/ermc/internal/manga/ui"

	"github.com/labstack/echo/v4"
)

func StartServer(port string) error {
	server := echo.New()

	MangaController.New(server)

	return server.Start(":"+port)
}
