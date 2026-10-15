package joinepubs_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/ismelen/inkomi/back/contracts/jobs"
	"github.com/ismelen/inkomi/back/epub-worker/internal/domain/ports"
	"github.com/ismelen/inkomi/back/epub-worker/internal/usecases/joinepubs"
	"github.com/ismelen/inkomi/back/epub-worker/internal/usecases/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type publishedEvent struct {
	subject string
	payload any
}

// subscribe runs Execute and returns the handler registered on the join subject.
func subscribe(t *testing.T, queue *mocks.MockQueue, store ports.JoinStore) ports.EventHandler {
	t.Helper()

	var handler ports.EventHandler
	queue.SubscribeFn = func(subject string, h ports.EventHandler) error {
		assert.Equal(t, string(jobs.SubjectJoin), subject)
		handler = h
		return nil
	}

	require.NoError(t, joinepubs.NewJoinEpubsUC(queue, store).Execute())
	require.NotNil(t, handler)
	return handler
}

func joinPayload(t *testing.T, parent *jobs.Parent) []byte {
	t.Helper()

	bytes, err := json.Marshal(jobs.Msg{Id: "child-1", UserId: 1, Parent: parent})
	require.NoError(t, err)
	return bytes
}

func TestJoinEpubsUC_Execute_InvalidJob_PublishesFailed(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		payload func(t *testing.T) []byte
	}{
		{"Invalid json", func(t *testing.T) []byte { return []byte("{") }},
		{"Without parent", func(t *testing.T) []byte { return joinPayload(t, nil) }},
		{"Without total", func(t *testing.T) []byte { return joinPayload(t, &jobs.Parent{Id: "parent-1"}) }},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			var events []publishedEvent
			queue := &mocks.MockQueue{
				PublishFn: func(ctx context.Context, subject string, payload any) error {
					events = append(events, publishedEvent{subject, payload})
					return nil
				},
			}
			addCalled := false
			store := &mocks.MockJoinStore{
				AddFn: func(ctx context.Context, parentId string, total int, msg jobs.Msg) ([]jobs.Msg, error) {
					addCalled = true
					return nil, nil
				},
			}
			handler := subscribe(t, queue, store)

			// Act
			err := handler(context.Background(), string(jobs.SubjectJoin), tc.payload(t))

			// Assert
			require.NoError(t, err)
			assert.False(t, addCalled)
			require.Len(t, events, 1)
			assert.Equal(t, string(jobs.SubjectFailed), events[0].subject)
			assert.NotEmpty(t, events[0].payload.(jobs.Msg).ErrorMsg)
		})
	}
}

func TestJoinEpubsUC_Execute_ValidJob_AddsToStore(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		stored []jobs.Msg
	}{
		{"Incomplete group", nil},
		{"Complete group", []jobs.Msg{{Id: "child-1"}, {Id: "child-2"}}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			published := false
			queue := &mocks.MockQueue{
				PublishFn: func(ctx context.Context, subject string, payload any) error {
					published = true
					return nil
				},
			}
			var gotParentId string
			var gotTotal int
			var gotMsg jobs.Msg
			store := &mocks.MockJoinStore{
				AddFn: func(ctx context.Context, parentId string, total int, msg jobs.Msg) ([]jobs.Msg, error) {
					gotParentId, gotTotal, gotMsg = parentId, total, msg
					return tc.stored, nil
				},
			}
			handler := subscribe(t, queue, store)

			// Act
			err := handler(context.Background(), string(jobs.SubjectJoin), joinPayload(t, &jobs.Parent{Id: "parent-1", Total: 2}))

			// Assert
			require.NoError(t, err)
			assert.False(t, published)
			assert.Equal(t, "parent-1", gotParentId)
			assert.Equal(t, 2, gotTotal)
			assert.Equal(t, "child-1", gotMsg.Id)
		})
	}
}

func TestJoinEpubsUC_Execute_StoreError_ReturnsError(t *testing.T) {
	t.Parallel()

	// Arrange
	storeErr := errors.New("store error")
	store := &mocks.MockJoinStore{
		AddFn: func(ctx context.Context, parentId string, total int, msg jobs.Msg) ([]jobs.Msg, error) {
			return nil, storeErr
		},
	}
	handler := subscribe(t, &mocks.MockQueue{}, store)

	// Act
	err := handler(context.Background(), string(jobs.SubjectJoin), joinPayload(t, &jobs.Parent{Id: "parent-1", Total: 2}))

	// Assert
	assert.ErrorIs(t, err, storeErr)
}
