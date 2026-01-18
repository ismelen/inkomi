package cli

import (
	"fmt"
	"ismelen/ermc/internal/api"
	MangaService "ismelen/ermc/internal/manga/logic"
	manga "ismelen/ermc/internal/manga/logic/models"
	"log"

	"github.com/spf13/cobra"
)

var opts = manga.Options{}

var rootCmd = &cobra.Command{
	Use:   "ermc",
	Short: "EReader Manga Converter (Go Port)",
	Long:  `A Go implementation of the Kindle Comic Converter to optimize comics and manga for e-readers.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := opts.ValidateAndNormalize(); err != nil {
			log.Fatal(err)
			return
		}
		
		links, err := MangaService.ProcessInputs(&opts)
		if err != nil {
			log.Fatal(err)
			return
		}

		fmt.Println("Outpus: ")
		for _, link := range links {
			fmt.Println(link)
		}
	},
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start API server",
	Run: func(cmd *cobra.Command, args []string) {
		port, _ := cmd.Flags().GetString("port")
		startServer(port)
	},
}

// startServer is a bridge to internal/api to avoid package cycle if any (though currently clean)
func startServer(port string) {
	if err := api.StartServer(port); err != nil {
		log.Fatal(err)
	}
}


func Execute() error {
	return rootCmd.Execute()
}

func init() {
	serveCmd.Flags().String("port", "8080", "Port to listen on")
	rootCmd.AddCommand(serveCmd)

	// rootCmd.PersistentFlags().StringVarP(&opts.Profile, "profile", "p", "KV", "Device profile (e.g., KV, KPW5)")
	// rootCmd.PersistentFlags().BoolVarP(&opts.Manga, "manga-style", "m", false, "Manga style (right-to-left reading)")
	// rootCmd.PersistentFlags().StringVarP(&opts.Output, "output", "o", "", "Output generated file to specified directory or file")
	// rootCmd.PersistentFlags().StringVarP(&opts.Title, "title", "t", "defaulttitle", "Comic title")
	// rootCmd.PersistentFlags().StringVarP(&opts.Format, "format", "f", "Auto", "Output format (Auto, EPUB, CBZ, PDF)")
	// rootCmd.PersistentFlags().BoolVar(&opts.LowRAM, "low-ram", false, "Enable low memory usage mode (slower)")

	rootCmd.PersistentFlags().StringSliceVarP(&opts.Inputs, "input", "i", []string{}, "Input files or directories (can be used multiple times)")

	// Metadatos y Archivos
	rootCmd.PersistentFlags().StringVarP(&opts.Title, "title", "t", "defaulttitle", "Comic title")
	rootCmd.PersistentFlags().StringVarP(&opts.Author, "author", "a", "Unknown", "Comic author")
	rootCmd.PersistentFlags().StringVarP(&opts.Format, "format", "f", "EPUB", "Output format (Auto, EPUB, CBZ, PDF)")
	rootCmd.PersistentFlags().Int64Var(&opts.TargetSize, "target-size", 200, "Target file size in MB")

	// Perfil y Rendimiento
	rootCmd.PersistentFlags().StringVarP(&opts.Profile, "profile", "p", "KoCC", "Device profile (e.g., KV, KPW5)")
	rootCmd.PersistentFlags().BoolVarP(&opts.LowRAM, "low-ram", "l", false, "Enable low memory usage mode (slower)")

	// Estilo y Lectura
	rootCmd.PersistentFlags().BoolVarP(&opts.Manga, "manga-style", "m", true, "Manga style (right-to-left reading)")
	rootCmd.PersistentFlags().BoolVar(&opts.SpreadShift, "spread-shift", true, "Shift double-page spreads")
	rootCmd.PersistentFlags().BoolVar(&opts.FileFusion, "file-fusion", true, "Combine multiple files into one")
	rootCmd.PersistentFlags().IntVar(&opts.SpreadSplitter, "spread-splitter", 2, "Spread splitter mode (0: Split, 1: Split+Rotated, 2: Rotated)")

	// Procesamiento de Imagen
	rootCmd.PersistentFlags().BoolVar(&opts.ColorMode, "color", true, "Enable color mode (keep colors)")
	rootCmd.PersistentFlags().IntVar(&opts.CroppingMode, "cropping", 2, "Cropping mode (0: No, 1: Margins, 2: Margins + page numbers)")
	rootCmd.PersistentFlags().Float32Var(&opts.CroppingPower, "cropping-power", 1.0, "Power of the cropping algorithm")
	rootCmd.PersistentFlags().BoolVar(&opts.RainbowEraser, "rainbow-eraser", true, "Enable rainbow filter eraser")
	rootCmd.PersistentFlags().BoolVar(&opts.ExtremBlackPoint, "black-point", true, "Enable extreme black point")
	rootCmd.PersistentFlags().IntVar(&opts.StretchUpscaleMode, "upscale-mode", 2, "Upscale mode (0: Nothing, 1: Stretching, 2: Upscaling)")
}
