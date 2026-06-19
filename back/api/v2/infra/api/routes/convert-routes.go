package routes

import (
	"ismelen/ermc/v2/infra/api/handlers"

	"github.com/go-chi/chi/v5"
)

// func SetupConvertRoutes(api *echo.Echo, handler *handlers.ConvertHandler) {
// 	api.POST("/convert", handler.Convert)
// 	api.GET("/status/{id}", handler.CheckStatus)
// 	api.GET("/download/{id}", handler.Download)
// 	api.PUT("/cancel/{id}", handler.Cancel)
// }

func SetupConvertRoutes(api *chi.Mux, handler *handlers.ConvertHandler) {
	api.Post("/convert", Wrap(handler.Convert))
	api.Get("/status/{id}", Wrap(handler.CheckStatus))
	api.Get("/download/{id}", Wrap(handler.Download))
	api.Put("/cancel/{id}", Wrap(handler.Cancel))
}
