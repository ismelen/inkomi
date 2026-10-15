package models

import "github.com/ismelen/inkomi/back/common/jobs"

// StatusFromSubject maps a job subject to the source status it represents.
func StatusFromSubject(s jobs.Subject) SourceStatus {
	switch s {
	case jobs.SubjectManga, jobs.SubjectLibrary:
		return SourceStatusQueued
	case jobs.SubjectStart:
		return SourceStatusProcessing
	case jobs.SubjectJoin:
		return SourceStatusJoining
	case jobs.SubjectJoined:
		return SourceStatusJoined
	case jobs.SubjectKepub:
		return SourceStatusKepubifying
	case jobs.SubjectDone:
		return SourceStatusDone
	case jobs.SubjectSent:
		return SourceStatusSent
	default:
		return SourceStatusFailed
	}
}
