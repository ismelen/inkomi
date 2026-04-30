package handlers

import (
	"ismelen/ermc/v2/domain"
	"ismelen/ermc/v2/infra/converters"
	"ismelen/ermc/v2/infra/crypto"
	"ismelen/ermc/v2/usecases"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/labstack/echo/v4"
)

type ConvertHandler struct{
	convert *usecases.ConvertMangaUC
	basePath string
}

func NewConvertHandler(convert *usecases.ConvertMangaUC) *ConvertHandler {
	tmp, err := os.MkdirTemp("", "ERMC(*)")
	if err != nil { log.Fatal(err) }
	
	return &ConvertHandler{
		convert: convert,
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

	chapters, err := converters.FormFilesToChapters(formFiles, filepath.Join(dstPath, "chapters"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()}) 
	}
	
	go func() {
		if err := ch.convert.Execute(chapters, req, dstPath); err != nil {
			//TODO: Send notification to user
			//TODO: Update state
			return
		}
		if req.Cloud {
			//TODO: Send and notify
		} else {
			//TODO: Notify
		}
		//TODO: Update state
	}()

	return c.JSON(http.StatusAccepted, echo.Map{"id": req.Id})
}

func (ch *ConvertHandler) CheckStatus(c echo.Context) error {
	return nil
}

func (ch *ConvertHandler) Download(c echo.Context) error {
	return nil
}

func (ch *ConvertHandler) Dispatch(c echo.Context) error {
	return nil
}

func (ch *ConvertHandler) Cancel(c echo.Context) error {
	return nil
}

