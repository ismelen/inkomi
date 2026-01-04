package MangaService

import (
	"fmt"
	EpubBuilder "ismelen/ermc/internal/manga/logic/builders/epub"
	MangaConstants "ismelen/ermc/internal/manga/logic/constants"
	PageConverter "ismelen/ermc/internal/manga/logic/coverters/page"
	manga "ismelen/ermc/internal/manga/logic/models"
	FileUtils "ismelen/ermc/internal/utils/file"
	ZipUtils "ismelen/ermc/internal/utils/zip"
	"os"
	"path/filepath"
	"runtime"

	"golang.org/x/sync/errgroup"
)

func ProcessInputs(opts *manga.Options) ([]string, error) {
	// Data Validation
	profileData, ok := MangaConstants.Profiles[opts.Profile]
	if !ok {
		return nil, fmt.Errorf("Unknown profile: %s", opts.Profile)
	}
	opts.ProfileData = profileData
	fmt.Printf("KCC Go Port running with profile: %s (%dx%d)\n", opts.Profile, opts.ProfileData.Width, opts.ProfileData.Height)

	if err := opts.ValidateAndNormalize(); err != nil {
		return nil, err
	}

	// Process images
	chaptersDir := filepath.Join(opts.Output, "chapters")
	defer func () {
		os.RemoveAll(chaptersDir)
	}()

	numCPUs := runtime.NumCPU()
	if opts.LowRAM {
		numCPUs = 1
	}
	var group errgroup.Group
	sem := make(chan struct{}, numCPUs)

	var pageNum int
	for _, chapter := range opts.InputData {
		if err := ExtractChapter(chapter, chaptersDir); err != nil {
			return nil, err
		}

		for _, page := range chapter.Pages {
			pageNum++;
			sem <- struct{}{}
			group.Go(func() error {
				defer func() { <- sem }()

				converter, err := PageConverter.New(page, opts, chapter.NormalizedName);
				if err != nil  {
					return err;
				}

				err = converter.Convert(pageNum)
				return err
			})
		}

		// sem <- struct{}{}

		// group.Go(func() error {
		// 	defer func(){ <- sem }()
		// 	var pageNum int
		// 	for _, page := range chapter.Pages {
		// 		converter, err := PageConverter.New(page, opts, chapter.NormalizedName);
		// 		if  err != nil {
		// 			return err
		// 		}

		// 		if err := converter.Convert(pageNum); err != nil {
		// 			return err
		// 		}
		// 	}
		// 	return nil
		// })
	}

	if err := group.Wait(); err != nil {
		return nil, err
	}

	// Generate volumes
	var resultPaths []string

	var targetSize int64 = 0
	if opts.FileFusion {
		targetSize = opts.TargetSize << 20
	}
	var (
		volSize int64
		lastIdx, volIdx int
	)
	inputsLen := len(opts.InputData)

	for idx, chapter := range opts.InputData {
		if volSize += chapter.Size; volSize < targetSize &&
			idx < inputsLen-1 {
			continue
		}

		filename := chapter.Title
		if opts.FileFusion {
			filename = fmt.Sprintf("%s Vol_%d", opts.Title, volIdx+1)
		}

		path, err := generateOutput(
			opts,
			filename,
			opts.InputData[lastIdx:idx+1]...,
		)
		if err != nil {
			return nil, err
		}
		resultPaths = append(resultPaths, path)
		volIdx++
		lastIdx = idx + 1
		volSize = 0
	}

	return resultPaths, nil
}

func ExtractChapter(chapter *manga.ChapterData, chaptersDir string) error {
	if !(FileUtils.HasCorrectExtension(chapter.Path, ".cbz")) {
		return fmt.Errorf("Cannot convert non [cbz] files")
	}
	newPath, pages, err := ZipUtils.UnzipFile(chapter.Path, chaptersDir, chapter.NormalizedName)
	if err != nil {
		return err
	}
	chapter.Path = newPath
	for _, page := range pages {
		chapter.Pages = append(chapter.Pages, manga.NewPageData(page))
	}

	return nil
}

func generateOutput(opts *manga.Options, dstFileName string, chapters ...*manga.ChapterData) (path string, err error) {
	switch opts.Format {
	case "Auto", "CBZ", "PDF", "EPUB":
		builder := EpubBuilder.New(opts, dstFileName, chapters...)
		path, err = builder.Build()
	}

	if path == "" {
		return "", fmt.Errorf("Cannot generate output for %s", dstFileName)
	}

	return
}

// func ProcessInputsOld(opts *manga.Options) ([]string, error) {
// 	profileData, ok := MangaConstants.Profiles[opts.Profile]
// 	if !ok {
// 		return nil, fmt.Errorf("Unknown profile: %s", opts.Profile)
// 	}
// 	opts.ProfileData = profileData
// 	fmt.Printf("KCC Go Port running with profile: %s (%dx%d)\n", opts.Profile, opts.ProfileData.Width, opts.ProfileData.Height)

// 	if err := opts.ValidateAndNormalize(); err != nil {
// 		return nil, err
// 	}

// 	// Extract chapter pages
// 	chaptersDir := filepath.Join(opts.Output, "chapters")
// 	for _, file := range opts.InputData {
// 		if !(FileUtils.HasCorrectExtension(file.Path, ".cbz")) {
// 			return nil, fmt.Errorf("Cannot convert non [cbz] files")
// 		}
// 		newPath, pages, err := ZipUtils.UnzipFile(file.Path, chaptersDir, file.NormalizedName)
// 		if err != nil {
// 			return nil, err
// 		}
// 		file.Path = newPath
// 		for _, page := range pages {
// 			file.Pages = append(file.Pages, manga.NewPageData(page))
// 		}
// 	}

// 	// Process each page
// 	for _, chapter := range opts.InputData {
// 		for _, page := range chapter.Pages {
// 			if err := PageConverter.New(page, opts); err != nil {
// 				return nil, err
// 			}
// 		}
// 	}

// 	var resultPaths []string

// 	// Generate volumes
// 	if !opts.FileFusion {
// 		for _, chapter := range opts.InputData {
// 			path, err := VolumeConverter.Convert(opts, chapter.Title, chapter)
// 			if err != nil {
// 				return nil, err
// 			}
// 			resultPaths = append(resultPaths, path)
// 		}
// 		return resultPaths, nil
// 	}

// 	targetSize := opts.TargetSize << 20 // bytes
// 	var volSize int64
// 	var lastIdx int
// 	var voldIdx int
// 	inputsLen := len(opts.InputData)

// 	for idx, chapter := range opts.InputData {
// 		if volSize += chapter.Size; volSize < targetSize &&
// 			idx < inputsLen-1 {
// 			continue
// 		}

// 		path, err := VolumeConverter.Convert(
// 			opts,
// 			fmt.Sprintf("%s Vol_%d", opts.Title, voldIdx+1),
// 			opts.InputData[lastIdx:idx+1]...,
// 		)
// 		if err != nil {
// 			return nil, err
// 		}
// 		resultPaths = append(resultPaths, path)
// 		voldIdx++
// 		lastIdx = idx + 1
// 		volSize = 0
// 	}

// 	return resultPaths, nil
// }
