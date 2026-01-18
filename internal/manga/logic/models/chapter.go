package manga

import (
	"fmt"
	FileUtils "ismelen/ermc/internal/utils/file"
	StringUtils "ismelen/ermc/internal/utils/strings"
	ZipUtils "ismelen/ermc/internal/utils/zip"
	"path/filepath"
	"strings"
)

type Chapter struct {
	Path string
	// Title          string
	NormalizedName string
	Pages          []*Page
}

func NewChapter(path string) *Chapter {
	name := filepath.Base(path)
	name = strings.TrimSuffix(name, filepath.Ext(name))
	return &Chapter{
		Path: path,
		// Title:          name,
		NormalizedName: StringUtils.NormalizeString(name),
		Pages:          []*Page{},
	}
}

func (this *Chapter) GetPages(dir string) ([]*Page, error) {
	hasCorrectExt := FileUtils.HasCorrectExtension(this.Path, ".cbz")
	if !hasCorrectExt {
		return nil, fmt.Errorf("Cannot convert non [cbz] files")
	}

	newPath, pages, err := ZipUtils.UnzipFile(
		this.Path,
		dir,
		this.NormalizedName,
	)
	if err != nil {
		return nil, err
	}

	this.Path = newPath
	for _, pagePath := range pages {
		this.Pages = append(this.Pages, NewPage(pagePath))
	}

	return this.Pages, nil
}
