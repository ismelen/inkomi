package convertmanga

import "github.com/ismelen/inkomi/back/common/jobs"

// NextSubject is the step that follows a converted manga.
func NextSubject(msg jobs.Msg) jobs.Subject {
	switch {
	case msg.Parent != nil:
		return jobs.SubjectJoin
	case msg.Kepubify:
		return jobs.SubjectKepub
	default:
		return jobs.SubjectDone
	}
}
