package manga

import (
	"fmt"
	"ismelen/ermc/internal/domain"
	"ismelen/ermc/internal/manga"
	"ismelen/ermc/internal/pkg"
	volumeBuilder "ismelen/ermc/internal/volume-builder"
	"log"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
)

var request = &CliConverterRequestDTO{}

var Cmd = &cobra.Command{
	Use:   "convert",
	Short: "The converter",
	Long:  "This converter will merge all chapters (if checked) in chunks of 200 MB",
	Run: func(cmd *cobra.Command, args []string) {
		start := time.Now()
		
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
		if err != nil {
			log.Fatal(err)
		}

		settings.SetImageSettings(domain.NewDefaultImageSettings())
		settings.SetVolumes(volumes)

		converter := manga.NewConverter(settings, int64(request.RamLimit))
		paths, err := converter.Convert(request.Format)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("Output:")
		for _, path := range paths {
			fmt.Println(path)
		}

		fmt.Printf("Time elapsed: %v\n", time.Since(start))
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
	Cmd.PersistentFlags().IntVarP(&request.RamLimit, "ram", "r", 3000, "Ram limit")
}

func getVolumes(dir string, settings *domain.Settings) ([]*domain.Volume, error) {
	files, err := pkg.GetChildsInfo(dir)

	var size int64
	for _, file := range files {
		size += file.Snd
	}
	fmt.Printf("Total processed: %v MB\n", size >> 20)
	
	if err != nil {
		log.Fatal(err)
	}

	fileExt := filepath.Ext(files[0].Fst)

	builder, err := volumeBuilder.GetBuilder(fileExt)
	if err != nil {
		log.Fatal(err)
	}

	return builder.FromPaths(settings, files...)
}
