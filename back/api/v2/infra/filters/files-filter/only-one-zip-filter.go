package filesFilter

import "path/filepath"

type OnlyOneZipFilter struct {
	FilesFilterBase
}

func (f *OnlyOneZipFilter) Filter(files []string) (bool, string) {
	if len(files) == 1 && filepath.Ext(files[0]) == ".zip" {
		return true, ".zip"
	}

	return f.Next(files)
}
