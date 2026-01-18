package MangaService

import (
	"fmt"
	EpubBuilder "ismelen/ermc/internal/manga/logic/builders/epub"
	PageConverter "ismelen/ermc/internal/manga/logic/coverters/page"
	manga "ismelen/ermc/internal/manga/logic/models"
	FileUtils "ismelen/ermc/internal/utils/file"
	StringUtils "ismelen/ermc/internal/utils/strings"
	ZipUtils "ismelen/ermc/internal/utils/zip"
	"os"
	"path/filepath"
	"runtime"

	"golang.org/x/sync/errgroup"
)

const PROCESSED_LOGGER_KEY = "Processed"

func ProcessInputs(opts *manga.Options) ([]string, error) {
	fmt.Printf("KCC Go Port running with profile: %s (%dx%d)\n", opts.Profile, opts.ProfileData.Width, opts.ProfileData.Height)

	// Process images
	chaptersDir := filepath.Join(opts.Output, "chapters")
	defer func () {
		os.RemoveAll(chaptersDir) 
	}()

	numCPUs := runtime.NumCPU() - runtime.NumGoroutine() - 1
	if numCPUs <= 0 {
		return nil, fmt.Errorf("Not enough threads")
	}
	if opts.LowRAM {
		numCPUs = 1
	}
	fmt.Printf("%d CPUs\n\n", numCPUs)

	var group errgroup.Group
	sem := make(chan struct{}, numCPUs)
	var pageNum int

	var resultPaths []string
	var volSize int64
	var lastIdx, volIdx int
	inputsLen := len(opts.InputData)
	
	for idx, chapter := range opts.InputData {
		if err := ExtractChapter(chapter, chaptersDir); err != nil {
			return nil, err
		}

		for _, page := range chapter.Pages {
			pageNum++;
			pNum := pageNum;
			cChapter := chapter
			sem <- struct{}{}
			group.Go(func() error {
				defer func() { <- sem }()

				converter, err := PageConverter.New(page, opts, cChapter.NormalizedName);
				if err != nil  {
					return err;
				}

				
				if err = converter.Convert(pNum); err != nil {
					return err
				}

				return nil
			})
		}

		
		// Montar volumen
		volSize += chapter.Size
		if volSize < opts.TargetSize &&
		idx < inputsLen-1 {
			continue
		}
		
		if err := group.Wait(); err != nil {
			return nil, err
		}

		filename := chapter.Title
		if opts.FileFusion {
			filename = fmt.Sprintf("%s Vol_%d", opts.Title, volIdx+1)
		}

		path, err := generateOutput(
			opts,
			StringUtils.NormalizeString(filename),
			opts.InputData[lastIdx:idx+1]...,
		)

		for i := range (idx+1-lastIdx) {
			opts.InputData[i] = nil
		}
		runtime.GC()

		if err != nil { return nil, err }
		resultPaths = append(resultPaths, path)
		volIdx++
		lastIdx = idx+1
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
