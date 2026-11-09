package models

import (
	"fmt"
	"path/filepath"
	"strings"
)

// BlockTitle names the epub merged from block, the index-th of total blocks of the parent folder.
func BlockTitle(parentTitle string, block []JoinItem, index, total int) string {
	first, firstOk := itemChapter(block[0])
	last, lastOk := itemChapter(block[len(block)-1])

	switch {
	case firstOk && lastOk && len(block) == 1:
		return fmt.Sprintf("%s Ch.%g", parentTitle, first)
	case firstOk && lastOk:
		return fmt.Sprintf("%s Ch.[%g - %g]", parentTitle, first, last)
	case total == 1:
		return parentTitle
	default:
		return fmt.Sprintf("%s Pt.%d", parentTitle, index+1)
	}
}

func itemChapter(item JoinItem) (float64, bool) {
	if n, ok := ExtractChapterNumber(item.Msg.Title); ok {
		return n, true
	}
	return ExtractChapterNumber(strings.TrimSuffix(item.Msg.Filename, filepath.Ext(item.Msg.Filename)))
}
