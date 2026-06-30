package validation

import (
	"ismelen/inkomi/internal/shared/filter"
	"path/filepath"
)

type OnlyOneZipFilter struct {
	filter.Base[[]string, string]
}

func (f *OnlyOneZipFilter) Filter(files []string) (bool, string) {
	hasZip := false
	for _, f := range files {
		ext := filepath.Ext(f)
		if ext == ".zip" {
			if hasZip || len(files) > 1 {
				return false, ""
			}
			hasZip = true
		}
	}
	return f.Next(files)
}
