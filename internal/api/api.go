package api

import (
	MangaController "ismelen/ermc/internal/manga/ui"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func StartServer(port string) error {
	server := echo.New()
	
	server.Use(middleware.RequestLogger())
	server.Use(middleware.CORS())


	MangaController.New(server)

	return server.Start(":"+port)
}
