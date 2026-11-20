package jobs

// Subject is a NATS subject of the job pipeline.
type Subject string

const (
	SubjectManga   Subject = "job.init.manga"
	SubjectLibrary Subject = "job.init.library"
	SubjectStart   Subject = "job.step.start"
	SubjectJoin    Subject = "job.step.join"
	SubjectJoined  Subject = "job.step.joined"
	SubjectKepub   Subject = "job.step.kepub"
	SubjectDone    Subject = "job.step.done"
	SubjectSent    Subject = "job.step.sent"
	SubjectFailed  Subject = "job.step.error"
)
