package downlolad

import (
	"ismelen/ermc/internal/pkg"
	"net/http"
	"os"
	"path/filepath"

	"github.com/labstack/echo/v4"
)

type Handler struct{}

func New(serv *echo.Echo) *Handler {
	handler := &Handler{}

	serv.GET("/download/:dir/:filename", handler.download)

	return handler
}

func (h *Handler) download(c echo.Context) error {
	dir := c.Param("dir")
	filename := c.Param("filename")

	path := filepath.Join(os.TempDir(), dir, filename)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return c.JSON(http.StatusNotFound, echo.Map{"error": "Archivo no encontrado"})
	}

	err := c.Attachment(path, filename)
	if err == nil {
		_ = os.Remove(path)
	}

	dirPath := filepath.Join(os.TempDir(), dir)
	if pkg.IsDirEmpty(dirPath) {
		os.RemoveAll(dirPath)
	}

	return nil
}
