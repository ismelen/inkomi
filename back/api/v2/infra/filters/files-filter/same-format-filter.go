package filesFilter

import "path/filepath"

type SameFormatFilter struct {
	FilesFilterBase
}

func (f *SameFormatFilter) Filter(files []string) (bool, string) {
	baseExt := filepath.Ext(files[0])
	for _, file := range files {
		ext := filepath.Ext(file)
		if baseExt != ext {
			return f.Next(files)
		}
	}

	return true, baseExt
}
