package validation

import (
	"ismelen/inkomi/internal/shared/filter"
	"path/filepath"
)

type SameFormatFilter struct {
	filter.Base[[]string, string]
}

func (f *SameFormatFilter) Filter(files []string) (bool, string) {
	if len(files) == 0 {
		return false, ""
	}

	baseExt := filepath.Ext(files[0])
	for _, file := range files {
		ext := filepath.Ext(file)
		if ext != baseExt {
			return false, ""
		}
	}

	return true, baseExt
}
