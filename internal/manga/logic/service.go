package MangaService

import (
	"fmt"
	MangaConstants "ismelen/ermc/internal/manga/logic/constants"
	PageConverter "ismelen/ermc/internal/manga/logic/coverters/page"
	VolumeConverter "ismelen/ermc/internal/manga/logic/coverters/volume"
	manga "ismelen/ermc/internal/manga/logic/models"
	FileUtils "ismelen/ermc/internal/utils/file"
	ZipUtils "ismelen/ermc/internal/utils/zip"
	"path/filepath"
)

func ProcessInputs(opts *manga.Options) ([]string, error) {
	profileData, ok := MangaConstants.Profiles[opts.Profile]
	if !ok {
		return nil, fmt.Errorf("Unknown profile: %s", opts.Profile)
	}
	opts.ProfileData = profileData
	fmt.Printf("KCC Go Port running with profile: %s (%dx%d)\n", opts.Profile, opts.ProfileData.Width, opts.ProfileData.Height)

	if err := opts.ValidateAndNormalize(); err != nil {
		return nil, err
	}

	// Extract chapter pages
	chaptersDir := filepath.Join(opts.Output, "chapters")
	for _, file := range opts.InputData {
		if !(FileUtils.HasCorrectExtension(file.Path, ".cbz")) {
			return nil, fmt.Errorf("Cannot convert non [cbz] files")
		}
		newPath, pages, err := ZipUtils.UnzipFile(file.Path, chaptersDir, file.NormalizedName)
		if err != nil {
			return nil, err
		}
		file.Path = newPath
		for _, page := range pages {
			file.Pages = append(file.Pages, manga.NewPageData(page))
		}
	}

	// Process each page
	for _, chapter := range opts.InputData {
		for _, page := range chapter.Pages {
			if err := PageConverter.Convert(page, opts); err != nil {
				return nil, err
			}
		}
	}

	var resultPaths []string

	// Generate volumes
	if !opts.FileFusion {
		for _, chapter := range opts.InputData {
			path, err := VolumeConverter.Convert(opts, chapter.Title, chapter)
			if err != nil {
				return nil, err
			}
			resultPaths = append(resultPaths, path)
		}
		return resultPaths, nil
	}

	targetSize := opts.TargetSize << 20 // bytes
	var volSize int64
	var lastIdx int
	var voldIdx int
	inputsLen := len(opts.InputData)

	for idx, chapter := range opts.InputData {
		if volSize += chapter.Size; volSize < targetSize &&
			idx < inputsLen-1 {
			continue
		}

		path, err := VolumeConverter.Convert(
			opts,
			fmt.Sprintf("%s Vol_%d", opts.Title, voldIdx+1),
			opts.InputData[lastIdx:idx+1]...,
		)
		if err != nil {
			return nil, err
		}
		resultPaths = append(resultPaths, path)
		voldIdx++
		lastIdx = idx + 1
		volSize = 0
	}

	return resultPaths, nil
}
