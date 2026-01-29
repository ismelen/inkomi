package manga

import (
	"fmt"
	documentBuilder "ismelen/ermc/internal/document-builder"
	"ismelen/ermc/internal/domain"
	"ismelen/ermc/internal/manga"
	volumeBuilder "ismelen/ermc/internal/volume-builder"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var request = &CliConverterRequestDTO{}

var Cmd = &cobra.Command{
	Use:   "convert",
	Short: "The converter",
	Long:  "This converter will merge all chapters (if checked) in chunks of 200 MB",
	Run: func(cmd *cobra.Command, args []string) {
		settings, err := domain.NewSettings(
			request.Author,
			request.Title,
			request.Profile,
			request.Merge,
			request.FirstVolumeNum,
		)
		if err != nil {
			log.Fatal(err)
		}

		volumes, err := getVolumes(request.InputDir, settings)

		settings.SetImageSettings(domain.NewDefaultImageSettings())
		settings.SetVolumes(volumes)

		documentBuilder, err := documentBuilder.GetBuilder(request.Format)
		if err != nil {
			log.Fatal(err)
		}
		documentBuilder.SetSettings(settings)

		converter := manga.NewConverter(settings, documentBuilder, int64(request.RamLimit))
		paths, err := converter.Convert()
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("Output:")
		for _, path := range paths {
			fmt.Println(path)
		}
	},
}

func init() {
	Cmd.PersistentFlags().StringVarP(&request.Author, "author", "a", "", "Manga author")
	Cmd.PersistentFlags().StringVarP(&request.Title, "title", "t", "", "Volume title")
	Cmd.PersistentFlags().StringVarP(&request.Profile, "profile", "p", "KoCC", "eReader model")
	Cmd.PersistentFlags().BoolVarP(&request.Merge, "merge", "m", true, "Merges all chapters")
	Cmd.PersistentFlags().IntVarP(&request.FirstVolumeNum, "firstVolumeNum", "n", 0, "Volume num to start counting")
	Cmd.PersistentFlags().StringVarP(&request.Format, "format", "f", "epub", "Result format")
	Cmd.PersistentFlags().StringVarP(&request.InputDir, "input", "i", "", "Input directory")
	Cmd.PersistentFlags().IntVarP(&request.RamLimit, "ram", "r", 8000, "Ram limit")
}

func getVolumes(dir string, settings *domain.Settings) ([]*domain.Volume, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}
	if len(files) == 0 {
		log.Fatal(fmt.Errorf("No files to convert"))
	}

	fileExt := filepath.Ext(files[0].Name())

	builder, err := volumeBuilder.GetBuilder(fileExt)
	if err != nil {
		log.Fatal(err)
	}

	return builder.FromPaths(settings, files...)
}
