package fromCbz

import (
	"fmt"
	"ismelen/ermc/internal/domain"
	"ismelen/ermc/internal/pkg"
	"mime/multipart"
	"os"
)

type VolumeFromCbz struct{}

func (v *VolumeFromCbz) FromMultipart(settings *domain.Settings, files ...*multipart.FileHeader) ([]*domain.Volume, error) {
	var volumes []*domain.Volume
	var size int64
	volIdx := settings.FirstVolumeNum
	var chapters []*domain.Chapter

	if volIdx == 0 {
		volIdx++
	}

	for idx, file := range files {
		size += file.Size
		isLast := idx >= len(files)-1

		pagePaths, err := pkg.UnzipFileForm(file, settings.Output.Chapters)
		if err != nil {
			return nil, err
		}

		var pages []*domain.Page
		for _, path := range pagePaths {
			pages = append(pages, domain.NewPage(path))
		}
		chapter := domain.NewChapter(file.Filename, file.Size, pages)
		chapters = append(chapters, chapter)

		if size < settings.TargetSize && !isLast {
			continue
		}

		if settings.Merge {
			volumes = append(volumes, domain.NewVolume(
				fmt.Sprintf("%s Vol_%d", settings.Title, volIdx),
				chapters...,
			))
		} else {
			volumes = append(volumes, domain.NewVolume(
				chapter.Name,
				chapters...,
			))
		}

		chapters = []*domain.Chapter{}
		volIdx++
		size = 0
	}

	return volumes, nil
}

func (v *VolumeFromCbz) FromPaths(settings *domain.Settings, files ...os.DirEntry) ([]*domain.Volume, error) {
	var volumes []*domain.Volume
	var size int64
	volIdx := settings.FirstVolumeNum
	var chapters []*domain.Chapter

	if volIdx == 0 {
		volIdx++
	}

	for idx, file := range files {
		info, err := file.Info()
		if err != nil {
			return nil, err
		}

		size += info.Size()
		isLast := idx >= len(files)-1

		pagePaths, err := pkg.UnizpFile(file, settings.Output.Chapters)
		if err != nil {
			return nil, err
		}

		var pages []*domain.Page
		for _, path := range pagePaths {
			pages = append(pages, domain.NewPage(path))
		}
		chapter := domain.NewChapter(file.Name(), info.Size(), pages)
		chapters = append(chapters, chapter)

		if size < settings.TargetSize && !isLast {
			continue
		}

		if settings.Merge {
			volumes = append(volumes, domain.NewVolume(
				fmt.Sprintf("%s Vol_%d", settings.Title, volIdx),
				chapters...,
			))
		} else {
			volumes = append(volumes, domain.NewVolume(
				chapter.Name,
				chapters...,
			))
		}

		chapters = []*domain.Chapter{}
		volIdx++
		size = 0
	}

	return volumes, nil
}
