package api

import (
	"ismelen/ermc/internal/ereader"
	MangaController "ismelen/ermc/internal/manga/ui"
	"net/http"
)

func StartServer(port string) error {
	MangaController.New()
	ereader.NewController()

	return http.ListenAndServe(":"+port, nil)
}
