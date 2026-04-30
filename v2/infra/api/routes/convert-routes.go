package routes

import (
	"ismelen/ermc/v2/infra/api/handlers"

	"github.com/labstack/echo/v4"
)

func SetupUserRoutes(api *echo.Echo, handler *handlers.ConvertHandler) {
	api.POST("/convert", handler.Convert)
	api.GET("/{id}/status", handler.CheckStatus)
	api.GET("/{id}/download", handler.Download)
	api.PUT("/{id}/output", handler.UpdateOutput)
	api.PUT("/{id}/cancel", handler.Cancel)
}