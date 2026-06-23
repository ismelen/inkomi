package routes

import (
	"fmt"
	"ismelen/inkomi/internal/domain"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/render"
)

func Wrap(f func(r *http.Request) (any, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, err := f(r)

		if err != nil {
			log.Println(err.Error())
			status := 500
			if apiErr, ok := err.(*domain.ApiError); ok {
				status = apiErr.Status
			}
			render.Status(r, status)
			render.JSON(w, r, map[string]any{"error": err.Error()})
		}

		if data == nil {
			render.NoContent(w, r)
			return
		}

		switch v := data.(type) {
		case domain.FileResponse:
			defer os.RemoveAll(v.Path)

			w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, v.Name))
			w.Header().Set("Content-Type", "application/octet-stream")
			http.ServeFile(w, r, v.Path)
		default:
			render.Status(r, http.StatusOK)
			render.JSON(w, r, v)
		}
	}
}
