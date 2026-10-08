package forwaredevents

/*
Events:
	- job.manga
	- job.library
	- job.step.epub.kepub
	- job.step.epub.join
	- job.step.done
	- job.step.sent
	- job.step.error
*/

type stepMsgDTO struct {
	Id         string `json:"id"`
	UserId     int    `json:"userId"`
	FolderId   string `json:"folderId,omitempty"`
	ShouldJoin bool   `json:"shouldJoin,omitempty"`
	Kepubify   bool   `json:"kepubify,omitempty"`
	ToCloud    bool   `json:"toCloud,omitempty"`
	Error      string `json:"error,omitempty"`
}
