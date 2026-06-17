package main

import (
	"ismelen/ermc/v2/infra/api/handlers"
	"ismelen/ermc/v2/infra/api/routes"
	"log"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	api := echo.New()
	api.Use(middleware.RequestLogger())
	api.Use(middleware.CORS())
	api.Use(middleware.BodyLimit("200M"))

	convertHandler := handlers.NewConvertHandler()
	routes.SetupConvertRoutes(api, convertHandler)

	if err := api.Start(":3000"); err != nil {
		log.Fatal(err)
	}
}
