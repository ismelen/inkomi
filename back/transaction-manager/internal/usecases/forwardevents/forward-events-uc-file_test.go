package forwaredevents_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/ismelen/inkomi/back/contracts/jobs"
	"github.com/ismelen/inkomi/back/transaction-manager/internal/domain/ports"
	forwaredevents "github.com/ismelen/inkomi/back/transaction-manager/internal/usecases/forwardevents"
	"github.com/ismelen/inkomi/back/transaction-manager/internal/usecases/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestForwardEventsUC_Execute_FileFields_UpdatesFile(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name         string
		job          jobs.Msg
		expectUpdate bool
	}{
		{"Filename and size", jobs.Msg{Id: "s-1", UserId: 1, Filename: "book.kepub.epub", Size: 200}, true},
		{"Neither", jobs.Msg{Id: "s-1", UserId: 1}, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			var handler ports.EventHandler
			queue := &mocks.MockQueue{SubscribeFn: func(subject jobs.Subject, h ports.EventHandler) error {
				handler = h
				return nil
			}}
			updated := false
			repo := &mocks.MockSourceRepository{
				UpdateFileFn: func(ctx context.Context, id string, userId int, filename string, size int64) error {
					updated = true
					assert.Equal(t, "book.kepub.epub", filename)
					assert.Equal(t, int64(200), size)
					return nil
				},
			}
			var sent map[string]any
			hub := &mocks.MockSocketHub{SendFn: func(id int, v any) error {
				sent = v.(map[string]any)
				return nil
			}}
			require.NoError(t, forwaredevents.NewForwardEventsUC(hub, queue, repo).Execute())
			payload, err := json.Marshal(tc.job)
			require.NoError(t, err)

			// Act
			err = handler(context.Background(), string(jobs.SubjectDone), payload)

			// Assert
			require.NoError(t, err)
			assert.Equal(t, tc.expectUpdate, updated)
			if tc.expectUpdate {
				assert.Equal(t, "book.kepub.epub", sent["filename"])
				assert.Equal(t, int64(200), sent["size"])
			} else {
				assert.NotContains(t, sent, "filename")
			}
		})
	}
}
