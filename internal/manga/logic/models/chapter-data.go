package manga

import (
	StringUtils "ismelen/ermc/internal/utils/strings"
	"path/filepath"
	"strings"
)

type ChapterData struct {
	Path           string
	Size           int64
	Title          string
	NormalizedName string
	Pages          []*PageData
}

func NewChapterData(path string, size int64) *ChapterData {
	name := filepath.Base(path)
	name = strings.TrimSuffix(name, filepath.Ext(name))
	return &ChapterData{
		Path:           path,
		Size:           size,
		Title:          name,
		NormalizedName: StringUtils.NormalizeString(name),
		Pages:          []*PageData{},
	}
}
