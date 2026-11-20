package handlers

import (
	"errors"
	"io"
	"log"
	"net/http"

	"github.com/ismelen/inkomi/back/api/internal/modules/library/domain/models"
	"github.com/ismelen/inkomi/back/api/internal/modules/library/usecases/getcover"
)

const coverCacheControl = "public, max-age=604800, immutable"

type CoversHandler struct {
	getCoverUC *getcover.GetCoverUC
}

func NewCoversHandler(getCoverUC *getcover.GetCoverUC) *CoversHandler {
	return &CoversHandler{getCoverUC: getCoverUC}
}

// HandleCover serves GET /books/cover?url=<mirror cover url>, streaming the image. It is public
// so <img> tags can load it; the source only fetches images from libgen hosts.
func (c *CoversHandler) HandleCover(w http.ResponseWriter, r *http.Request) {
	cover, err := c.getCoverUC.Execute(r.Context(), r.URL.Query().Get("url"))
	switch {
	case errors.Is(err, models.ErrInvalidCoverURL):
		http.Error(w, "cover not found", http.StatusNotFound)
		return
	case err != nil:
		log.Printf("%s %s: %v", r.Method, r.URL.Path, err)
		http.Error(w, "cover unavailable", http.StatusBadGateway)
		return
	}
	defer cover.Body.Close()

	w.Header().Set("Content-Type", cover.ContentType)
	w.Header().Set("Cache-Control", coverCacheControl)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	if _, err := io.Copy(w, cover.Body); err != nil {
		log.Printf("%s %s: stream cover: %v", r.Method, r.URL.Path, err)
	}
}
