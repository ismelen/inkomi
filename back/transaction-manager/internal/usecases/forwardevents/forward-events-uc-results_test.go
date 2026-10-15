package forwaredevents_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/ismelen/inkomi/back/common/jobs"
	"github.com/ismelen/inkomi/back/transaction-manager/internal/domain/models"
	"github.com/ismelen/inkomi/back/transaction-manager/internal/domain/ports"
	forwaredevents "github.com/ismelen/inkomi/back/transaction-manager/internal/usecases/forwardevents"
	"github.com/ismelen/inkomi/back/transaction-manager/internal/usecases/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type parentUpdate struct {
	status   models.SourceStatus
	errorMsg string
}

// resultRepo stores upserted results in memory and records the updates applied to the parent.
func resultRepo(t *testing.T, existing []models.Source, parent *parentUpdate) *mocks.MockSourceRepository {
	t.Helper()

	results := append([]models.Source{}, existing...)
	return &mocks.MockSourceRepository{
		UpsertResultFn: func(ctx context.Context, result *models.Source) error {
			results = append(results, *result)
			return nil
		},
		GetByIdAndUserIdFn: func(ctx context.Context, id string, userId int) (*models.Source, error) {
			for i := range results {
				if results[i].Id == id {
					return &results[i], nil
				}
			}
			return nil, models.ErrSourceNotFound
		},
		ListResultsByParentIdAndUserIdFn: func(ctx context.Context, parentId string, userId int) ([]models.Source, error) {
			return results, nil
		},
		UpdateStatusFn: func(ctx context.Context, id string, userId int, status models.SourceStatus) error {
			assert.Equal(t, "folder-1", id)
			parent.status = status
			return nil
		},
		UpdateErrorMsgFn: func(ctx context.Context, id string, userId int, errorMsg string) error {
			assert.Equal(t, "folder-1", id)
			parent.status, parent.errorMsg = models.SourceStatusFailed, errorMsg
			return nil
		},
	}
}

func handleResult(t *testing.T, repo ports.ISourceRepository, subject jobs.Subject, job jobs.Msg) []map[string]any {
	t.Helper()

	var handler ports.EventHandler
	queue := &mocks.MockQueue{
		SubscribeFn: func(subject jobs.Subject, h ports.EventHandler) error {
			handler = h
			return nil
		},
	}
	var sent []map[string]any
	hub := &mocks.MockSocketHub{
		SendFn: func(id int, v any) error {
			sent = append(sent, v.(map[string]any))
			return nil
		},
	}
	require.NoError(t, forwaredevents.NewForwardEventsUC(hub, queue, repo).Execute())

	payload, err := json.Marshal(job)
	require.NoError(t, err)
	require.NoError(t, handler(context.Background(), string(subject), payload))
	return sent
}

func resultJob(id string, index, total int) jobs.Msg {
	return jobs.Msg{
		Id:       id,
		UserId:   1,
		Title:    "Folder Ch.[1 - 2]",
		Filename: "Folder Ch.[1 - 2].epub",
		Size:     100,
		Result:   &jobs.Result{ParentId: "folder-1", Index: index, Total: total},
	}
}

func finishedResult(id string, status models.SourceStatus) models.Source {
	return models.Source{Id: id, UserId: 1, Status: status, IsResult: true}
}

func TestForwardEventsUC_Execute_Result_UpsertsAndSendsResultData(t *testing.T) {
	t.Parallel()

	// Arrange
	var upserted *models.Source
	parent := &parentUpdate{}
	repo := resultRepo(t, nil, parent)
	upsert := repo.UpsertResultFn
	repo.UpsertResultFn = func(ctx context.Context, result *models.Source) error {
		upserted = result
		return upsert(ctx, result)
	}

	// Act
	sent := handleResult(t, repo, jobs.SubjectKepub, resultJob("result-0", 0, 2))

	// Assert
	require.NotNil(t, upserted)
	assert.True(t, upserted.IsResult)
	assert.Equal(t, "folder-1", *upserted.ParentId)
	assert.Equal(t, models.SourceStatusKepubifying, upserted.Status)
	assert.Equal(t, int64(100), *upserted.Size)

	require.Len(t, sent, 1)
	assert.Equal(t, "result-0", sent[0]["id"])
	assert.Equal(t, models.SourceStatusKepubifying, sent[0]["status"])
	assert.Equal(t, "folder-1", *sent[0]["result"].(*models.Source).ParentId)
	assert.Empty(t, parent.status)
}

func TestForwardEventsUC_Execute_LastResult_UpdatesParent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name           string
		existing       []models.Source
		subject        jobs.Subject
		expectedStatus models.SourceStatus
		expectedError  string
	}{
		{
			name:     "Other result unfinished",
			existing: []models.Source{finishedResult("result-0", models.SourceStatusKepubifying)},
			subject:  jobs.SubjectDone,
		},
		{
			name:           "All done",
			existing:       []models.Source{finishedResult("result-0", models.SourceStatusSent)},
			subject:        jobs.SubjectDone,
			expectedStatus: models.SourceStatusDone,
		},
		{
			name:           "One failed",
			existing:       []models.Source{finishedResult("result-0", models.SourceStatusDone)},
			subject:        jobs.SubjectFailed,
			expectedStatus: models.SourceStatusFailed,
			expectedError:  "1 of 2 results failed",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			parent := &parentUpdate{}
			repo := resultRepo(t, tc.existing, parent)

			// Act
			sent := handleResult(t, repo, tc.subject, resultJob("result-1", 1, 2))

			// Assert
			assert.Equal(t, tc.expectedStatus, parent.status)
			assert.Equal(t, tc.expectedError, parent.errorMsg)
			if tc.expectedStatus == "" {
				assert.Len(t, sent, 1)
				return
			}
			require.Len(t, sent, 2)
			assert.Equal(t, "folder-1", sent[1]["id"])
			assert.Equal(t, tc.expectedStatus, sent[1]["status"])
		})
	}
}
