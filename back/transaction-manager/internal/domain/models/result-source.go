package models

import "github.com/ismelen/inkomi/back/contracts/jobs"

// NewResultSource builds the source row of a join result from the job that reports it.
func NewResultSource(job jobs.Msg, status SourceStatus) *Source {
	result := &Source{
		Id:       job.Id,
		UserId:   job.UserId,
		Title:    job.Title,
		Filename: job.Filename,
		Kepubify: job.Kepubify,
		ToCloud:  job.ToCloud,
		Status:   status,
		IsFile:   true,
		IsResult: true,
	}

	if job.Result != nil {
		result.ParentId = &job.Result.ParentId
		result.Position = job.Result.Index
	}
	if job.Size > 0 {
		result.Size = &job.Size
	}
	if job.ErrorMsg != "" {
		result.ErrorMsg = &job.ErrorMsg
	}

	return result
}
