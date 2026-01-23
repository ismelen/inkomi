package MangaController

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
)

func conversionStatus(c echo.Context, statuses *sync.Map) error {
	c.Response().Header().Set(echo.HeaderContentType, "text/event-stream")
	c.Response().Header().Set("Cache-Control", "no-cache")
	c.Response().Header().Set("Connection", "keep-alive")
	c.Response().WriteHeader(http.StatusOK)

	id := c.Param("id")
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	ctx := c.Request().Context()

	for {
		select {
		case <-ctx.Done():
			{
				return nil
			}
		case <-ticker.C:
			value, ok := statuses.Load(id)
			if !ok || value == nil {
				data, _ := json.Marshal(statusData{
					Err: "ID not found",
				})
				fmt.Fprintf(c.Response(), "data: %s\n\n", data)
				c.Response().Flush()
				return nil
			}
			status := value.(*statusData)
			data, _ := json.Marshal(status)
			if len(status.Paths) > 0 {
				statuses.Delete(id)
			}

			fmt.Fprintf(c.Response(), "data: %s\n\n", data)
			c.Response().Flush()
			status.Msg = ""
		}
	}
}
