package kepubify

import (
	"archive/zip"
	"context"
	"fmt"
	"ismelen/ermc/internal/cloud"
	"ismelen/ermc/internal/pkg"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/pgaskin/kepubify/v4/kepub"
)

type Handler struct{}

func NewHandler(serv *echo.Echo) *Handler {
	handler := &Handler{}

	serv.POST("/kepubify", handler.handleRequest)

	return handler
}

func (h *Handler) handleRequest(c echo.Context) error {
	dto := new(KepubifyRequestDTO)
	if err := c.Bind(dto); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	files, err := getFiles(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	cloudType := "google"
	if dto.CloudToken == "" {
		cloudType = "local"
	}
	cloudService, err := cloud.GetCloud(cloudType)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	cloudService.Init(dto.CloudToken, dto.CloudFolder)

	dst, err := pkg.NewTempDir("ermck")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	var paths []string

	for _, file := range files {
		path, err := kepubify(file, dst)
		if err != nil {
			continue
		}
		newPath, err := cloudService.Upload(path)
		if err != nil {
			continue
		}
		paths = append(paths, newPath)
	}

	return c.JSON(http.StatusOK, echo.Map{"paths": paths})
}

func getFiles(c echo.Context) ([]*multipart.FileHeader, error) {
	form, err := c.MultipartForm()
	if err != nil {
		return nil, err
	}

	files := form.File["files"]
	if len(files) == 0 {
		return nil, fmt.Errorf("nothing to convert")
	}

	return files, nil
}

func kepubify(file *multipart.FileHeader, dst string) (string, error) {
	ext := filepath.Ext(file.Filename)
	noExtName := strings.TrimSuffix(file.Filename, ext)
	kpath := filepath.Join(dst, noExtName+".kepub.epub")

	out, err := os.Create(kpath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	tempFile, err := os.Create(filepath.Join(dst, file.Filename))
	if err != nil {
		return "", err
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	in, err := zip.OpenReader(tempFile.Name())
	if err != nil {
		return "", err
	}
	defer in.Close()

	converter := kepub.NewConverter()
	ctx := context.Background()

	return kpath, converter.Convert(ctx, out, in)
}
