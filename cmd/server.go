package cmd

import (
	manga "ismelen/ermc/internal/manga/http"
	"log"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/spf13/cobra"
)

var serverCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start API Server",
	Run: func(cmd *cobra.Command, args []string) {
		port, _ := cmd.Flags().GetString("port")
		if err := startServer(port); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	serverCmd.Flags().String("port", "8080", "Port to listen on")
}

func startServer(port string) error {
	server := echo.New()

	server.Use(middleware.RequestLogger())
	server.Use(middleware.CORS())

	manga.NewHandler(server)

	return server.Start(":" + port)
}
