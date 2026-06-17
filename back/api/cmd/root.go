package cmd

import (
	manga "ismelen/ermc/internal/manga/cli"

	"github.com/spf13/cobra"
)

func Execute() error {
	return cmd.Execute()
}

var cmd = &cobra.Command{
	Use:   "ermc",
	Short: "EReader Manga Converter",
}

func init() {
	cmd.AddCommand(manga.Cmd)
	cmd.AddCommand(serverCmd)
}
