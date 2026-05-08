package routes

import (
	"ismelen/ermc/v2/infra/api/handlers"

	"github.com/labstack/echo/v4"
)

func SetupConvertRoutes(api *echo.Echo, handler *handlers.ConvertHandler) {
	api.POST("/convert", handler.Convert)
	api.GET("/status/{id}", handler.CheckStatus)
	api.GET("/download/{id}", handler.Download)
	api.PUT("/output/{id}", handler.Dispatch)
	api.PUT("/cancel/{id}", handler.Cancel)
}