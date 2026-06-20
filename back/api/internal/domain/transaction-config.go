package domain

import (
	"fmt"
	"ismelen/inkomi/internal/infra/helpers"
)

type TransactionConfig struct {
	Author      string `form:"author,omitempty"`
	Title       string `form:"title,omitempty"`
	Profile     string `form:"profile"`
	Merge       bool   `form:"merge,omitempty"`
	Id          string
	Cloud       bool   `form:"cloud"`
	CloudToken  string `form:"cloud_token,omitempty"`
	CloudFolder string `form:"cloud_folder,omitempty"`
	NotifyToken string `form:"notify_token,omitempty"`
}

func (t *TransactionConfig) WithId(id string) *TransactionConfig {
	trans := *t
	trans.Id = id
	return &trans
}

func (t *TransactionConfig) UpdateTitle(chapters []*Chapter) {
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

	fstChNum, fstOk := helpers.ExtractChapterNumber(fstChName)
	lastChNum, lastOk := helpers.ExtractChapterNumber(lastChName)

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
