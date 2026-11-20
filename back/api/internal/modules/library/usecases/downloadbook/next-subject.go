package downloadbook

import (
	"strings"

	"github.com/ismelen/inkomi/back/api/internal/platform/jobs"
)

// nextSubject routes a downloaded book like the transactions module routes an uploaded file:
// only epubs are joined or kepubified, anything else is done.
func nextSubject(msg jobs.Msg) jobs.Subject {
	lower := strings.ToLower(msg.Filename)
	switch {
	case !strings.HasSuffix(lower, ".epub"):
		return jobs.SubjectDone
	case msg.Parent != nil:
		return jobs.SubjectJoin
	case msg.Kepubify && !strings.HasSuffix(lower, ".kepub.epub"):
		return jobs.SubjectKepub
	default:
		return jobs.SubjectDone
	}
}
