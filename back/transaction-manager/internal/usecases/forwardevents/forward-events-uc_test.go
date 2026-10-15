package forwaredevents_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/ismelen/inkomi/back/contracts/jobs"
	"github.com/ismelen/inkomi/back/transaction-manager/internal/domain/models"
	"github.com/ismelen/inkomi/back/transaction-manager/internal/domain/ports"
	forwaredevents "github.com/ismelen/inkomi/back/transaction-manager/internal/usecases/forwardevents"
	"github.com/ismelen/inkomi/back/transaction-manager/internal/usecases/mocks"
	"github.com/stretchr/testify/assert"
)

func TestForwardEventsUC_Execute_SuccessfullyProcessesEventAndUpdatesStatus(t *testing.T) {
	// Arrange
	var capturedHandler ports.EventHandler
	mockQueue := &mocks.MockQueue{
		SubscribeFn: func(subject jobs.Subject, handler ports.EventHandler) error {
			capturedHandler = handler
			return nil
		},
	}

	var updatedStatus models.SourceStatus
	mockRepo := &mocks.MockSourceRepository{
		UpdateStatusFn: func(ctx context.Context, id string, userId int, status models.SourceStatus) error {
			updatedStatus = status
			return nil
		},
	}

	var sentMsg any
	mockHub := &mocks.MockSocketHub{
		SendFn: func(id int, v any) error {
			sentMsg = v
			return nil
		},
	}

	uc := forwaredevents.NewForwardEventsUC(mockHub, mockQueue, mockRepo)

	stepMsg := jobs.Msg{
		Id:     "step-123",
		UserId: 1,
	}
	payload, _ := json.Marshal(stepMsg)

	// Act
	uc.Execute()
	err := capturedHandler(context.Background(), string(jobs.SubjectKepub), payload)

	// Assert
	assert.NotNil(t, capturedHandler, "Expected queue.Subscribe to be called with a handler")
	assert.NoError(t, err)
	assert.Equal(t, models.SourceStatusKepubifying, updatedStatus)

	expectedMsg := map[string]any{
		"id":     "step-123",
		"status": models.SourceStatusKepubifying,
	}
	assert.Equal(t, expectedMsg, sentMsg)
}

func TestForwardEventsUC_Execute_ProcessesEventWithErrorMessage(t *testing.T) {
	// Arrange
	var capturedHandler ports.EventHandler
	mockQueue := &mocks.MockQueue{
		SubscribeFn: func(subject jobs.Subject, handler ports.EventHandler) error {
			capturedHandler = handler
			return nil
		},
	}

	var storedErrorMsg string
	mockRepo := &mocks.MockSourceRepository{
		UpdateStatusFn: func(ctx context.Context, id string, userId int, status models.SourceStatus) error {
			t.Fatal("UpdateStatus should not be called when the job carries an error")
			return nil
		},
		UpdateErrorMsgFn: func(ctx context.Context, id string, userId int, errorMsg string) error {
			storedErrorMsg = errorMsg
			return nil
		},
	}

	var sentMsg any
	mockHub := &mocks.MockSocketHub{
		SendFn: func(id int, v any) error {
			sentMsg = v
			return nil
		},
	}

	uc := forwaredevents.NewForwardEventsUC(mockHub, mockQueue, mockRepo)

	stepMsg := jobs.Msg{
		Id:       "step-123",
		UserId:   1,
		ErrorMsg: "something went wrong",
	}
	payload, _ := json.Marshal(stepMsg)

	// Act
	uc.Execute()
	err := capturedHandler(context.Background(), string(jobs.SubjectFailed), payload)

	// Assert
	assert.NotNil(t, capturedHandler)
	assert.NoError(t, err)
	assert.Equal(t, "something went wrong", storedErrorMsg)

	expectedMsg := map[string]any{
		"id":     "step-123",
		"status": models.SourceStatusFailed,
		"error":  "something went wrong",
	}
	assert.Equal(t, expectedMsg, sentMsg)
}

func TestForwardEventsUC_Execute_InvalidJsonPayload_ReturnsError(t *testing.T) {
	// Arrange
	var capturedHandler ports.EventHandler
	mockQueue := &mocks.MockQueue{
		SubscribeFn: func(subject jobs.Subject, handler ports.EventHandler) error {
			capturedHandler = handler
			return nil
		},
	}

	uc := forwaredevents.NewForwardEventsUC(
		&mocks.MockSocketHub{},
		mockQueue,
		&mocks.MockSourceRepository{},
	)

	// Act
	uc.Execute()
	err := capturedHandler(context.Background(), "job.step.start", []byte("invalid-json"))

	// Assert
	assert.NotNil(t, capturedHandler)
	assert.Error(t, err)
}

func TestForwardEventsUC_Execute_RepoUpdateStatusReturnsError_ReturnsError(t *testing.T) {
	// Arrange
	var capturedHandler ports.EventHandler
	mockQueue := &mocks.MockQueue{
		SubscribeFn: func(subject jobs.Subject, handler ports.EventHandler) error {
			capturedHandler = handler
			return nil
		},
	}

	repoErr := errors.New("db error")
	mockRepo := &mocks.MockSourceRepository{
		UpdateStatusFn: func(ctx context.Context, id string, userId int, status models.SourceStatus) error {
			return repoErr
		},
	}

	mockHub := &mocks.MockSocketHub{
		SendFn: func(id int, v any) error {
			t.Fatal("Send should not be called if UpdateStatus fails")
			return nil
		},
	}

	uc := forwaredevents.NewForwardEventsUC(mockHub, mockQueue, mockRepo)

	stepMsg := jobs.Msg{
		Id:     "step-123",
		UserId: 1,
	}
	payload, _ := json.Marshal(stepMsg)

	// Act
	uc.Execute()
	err := capturedHandler(context.Background(), "job.step.start", payload)

	// Assert
	assert.NotNil(t, capturedHandler)
	assert.ErrorIs(t, err, repoErr)
}
