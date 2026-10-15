package models

import "github.com/ismelen/inkomi/back/common/jobs"

// JoinItem is a child of a join folder together with the size of its file in cloud storage.
type JoinItem struct {
	Msg  jobs.Msg
	Size int64
}
