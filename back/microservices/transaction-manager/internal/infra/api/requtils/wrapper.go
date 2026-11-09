package requtil

import (
	"log"
	"net/http"

	"github.com/go-chi/render"
)

func Wrap[R any](f func(w http.ResponseWriter, r *http.Request) (*R, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, err := f(w, r)

		if err != nil {
			log.Println(err.Error())
			status := 500
			if apiErr, ok := err.(*ApiError); ok {
				status = apiErr.Status
			}
			render.Status(r, status)
			render.JSON(w, r, map[string]any{"error": err.Error()})
			return
		}

		if data == nil {
			render.NoContent(w, r)
			return
		}

		render.Status(r, http.StatusOK)
		render.JSON(w, r, data)
	}
}
