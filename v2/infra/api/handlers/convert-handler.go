package handlers

import (
	"ismelen/ermc/v2/domain"
	"ismelen/ermc/v2/infra/converters"
	"ismelen/ermc/v2/infra/crypto"
	"ismelen/ermc/v2/infra/state"
	"ismelen/ermc/v2/usecases"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/labstack/echo/v4"
)

type ConvertHandler struct{
	basePath string
}

func NewConvertHandler() *ConvertHandler {
	tmp, err := os.MkdirTemp("", "ERMC(*)")
	if err != nil { log.Fatal(err) }
	
	return &ConvertHandler{
		basePath: tmp,
	}
}

func (ch *ConvertHandler) Convert(c echo.Context) error {
	req := new(domain.ConvertConfig)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()}) 
	}

	formFiles, err := GetFormFiles(c, "files")
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()}) 
	}

	req.Id = crypto.GetRandomID(6)
	dstPath := filepath.Join(ch.basePath, req.Id)

	chapters, err := converters.FilesToChapters(formFiles, filepath.Join(dstPath, "chapters"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()}) 
	}

	if req.Title == "" {
		req.Title = filepath.Base(chapters[0].Path)
	}
	
	converter := usecases.NewConvertMangaUC()
	go converter.Execute(chapters, req, dstPath)
	return c.JSON(http.StatusAccepted, echo.Map{"id": req.Id, "title": req.Title})
}

func (ch *ConvertHandler) CheckStatus(c echo.Context) error {
	id := c.Param("id")
	stateMng := state.GetManager()

	processed, err := stateMng.CheckProgress(id)
	if err != nil {
		log.Println(err.Error())
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, echo.Map{"progress": processed})
}

func (ch *ConvertHandler) Download(c echo.Context) error {
	id := c.Param("id")
	stateMng := state.GetManager()

	path, err := stateMng.GetResultPath(id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()}) 
	}

	return c.File(path)
}

func (ch *ConvertHandler) Cancel(c echo.Context) error {
	id := c.Param(("id"))
	stateMng := state.GetManager()

	stateMng.DeleteTransaction(id)

	return nil
}

