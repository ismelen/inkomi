package convert

import (
	"fmt"
	"ismelen/inkomi/internal/domain/manga"
	"ismelen/inkomi/internal/shared/strutil"
)

type TransactionConfig struct {
	Author      string
	Title       string
	Profile     string
	Merge       bool
	Id          string
	Cloud       bool
	CloudToken  string
	CloudFolder string
	NotifyToken string
	ProfileData *manga.Profile
}

func (t *TransactionConfig) WithId(id string) *TransactionConfig {
	trans := *t
	trans.Id = id
	return &trans
}

func (t *TransactionConfig) UpdateTitle(chapters []*manga.Chapter) {
	if !t.Merge && t.Title == "" {
		t.Title = chapters[0].Filename
		return
	}

	fstChName := chapters[0].Filename
	lastChName := chapters[len(chapters)-1].Filename

	if t.Title == "" {
		if len(chapters) == 1 {
			t.Title = fstChName
			return
		}
		t.Title = fmt.Sprintf("%s - %s", fstChName, lastChName)
		return
	}

	fstChNum, fstOk := strutil.ExtractChapterNumber(fstChName)
	lastChNum, lastOk := strutil.ExtractChapterNumber(lastChName)

	if !fstOk || !lastOk {
		if len(chapters) == 1 {
			t.Title = fstChName
			return
		}
		t.Title = fmt.Sprintf("%s - %s", fstChName, lastChName)
		return
	}

	if len(chapters) == 1 {
		t.Title = fmt.Sprintf("%s Ch.%g", t.Title, fstChNum)
		return
	}
	t.Title = fmt.Sprintf("%s Ch.[%g - %g]", t.Title, fstChNum, lastChNum)
}
