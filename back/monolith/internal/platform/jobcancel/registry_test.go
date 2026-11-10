package jobcancel_test

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/ismelen/inkomi/back/monolith/internal/platform/jobcancel"
	"github.com/ismelen/inkomi/back/monolith/internal/platform/jobrun"
	"github.com/ismelen/inkomi/back/monolith/internal/platform/jobs"
	"github.com/ismelen/inkomi/back/monolith/internal/platform/sqlitedb"
	"github.com/ismelen/inkomi/back/monolith/migrations"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type publisher struct{ subjects []string }

func (p *publisher) Publish(ctx context.Context, subject string, payload any) error {
	p.subjects = append(p.subjects, subject)
	return nil
}

func newRegistry(t *testing.T, path string) *jobcancel.Registry {
	t.Helper()
	db, err := sqlitedb.Open(path, migrations.Joins)
	require.NoError(t, err)
	t.Cleanup(func() {
		// the database lives in a temp dir
		_ = db.Close()
	})
	r, err := jobcancel.NewRegistry(db, time.Now)
	require.NoError(t, err)
	return r
}

func delivery(t *testing.T, msg jobs.Msg) jobrun.Delivery {
	t.Helper()
	payload, err := json.Marshal(msg)
	require.NoError(t, err)
	return jobrun.Delivery{Subject: string(jobs.SubjectManga), Payload: payload, LastAttempt: true}
}

func TestRegistry_Cancel_SurvivesRestart(t *testing.T) {
	t.Parallel()

	// Arrange
	path := filepath.Join(t.TempDir(), "jobs.db")
	first := newRegistry(t, path)

	// Act
	require.NoError(t, first.Cancel(context.Background(), []string{"a", "b"}))
	restarted := newRegistry(t, path)

	// Assert
	assert.True(t, restarted.IsCancelled("x", "b"))
	assert.False(t, restarted.IsCancelled("x", ""))
}

func TestHandle_CancelledBefore_SkipsWithoutFailure(t *testing.T) {
	t.Parallel()

	// Arrange: a chapter of a cancelled folder
	r := newRegistry(t, filepath.Join(t.TempDir(), "jobs.db"))
	require.NoError(t, r.Cancel(context.Background(), []string{"folder"}))
	pub := &publisher{}
	ran := false
	h := jobrun.Handle(pub, func(ctx context.Context, msg jobs.Msg, d jobrun.Delivery) error {
		ran = true
		return nil
	}, jobrun.WithCancellations(r))

	// Act
	err := h(context.Background(), delivery(t, jobs.Msg{Id: "ch-1", Parent: &jobs.Parent{Id: "folder", Total: 2}}))

	// Assert
	require.NoError(t, err)
	assert.False(t, ran)
	assert.Empty(t, pub.subjects)
}

func TestHandle_CancelledWhileRunning_InterruptsWithoutFailure(t *testing.T) {
	t.Parallel()

	// Arrange
	r := newRegistry(t, filepath.Join(t.TempDir(), "jobs.db"))
	pub := &publisher{}
	started := make(chan struct{})
	h := jobrun.Handle(pub, func(ctx context.Context, msg jobs.Msg, d jobrun.Delivery) error {
		close(started)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(5 * time.Second):
			return errors.New("not interrupted")
		}
	}, jobrun.WithCancellations(r))

	// Act
	done := make(chan error, 1)
	go func() { done <- h(context.Background(), delivery(t, jobs.Msg{Id: "src-1"})) }()
	<-started
	require.NoError(t, r.Cancel(context.Background(), []string{"src-1"}))

	// Assert
	select {
	case err := <-done:
		require.NoError(t, err)
		assert.Empty(t, pub.subjects, "a cancelled job doesn't report a failure")
	case <-time.After(3 * time.Second):
		t.Fatal("the job wasn't interrupted")
	}
}

func TestHandle_FailureNotCancelled_StillReported(t *testing.T) {
	t.Parallel()

	// Arrange
	r := newRegistry(t, filepath.Join(t.TempDir(), "jobs.db"))
	pub := &publisher{}
	h := jobrun.Handle(pub, func(ctx context.Context, msg jobs.Msg, d jobrun.Delivery) error {
		return errors.New("boom")
	}, jobrun.WithCancellations(r))

	// Act
	err := h(context.Background(), delivery(t, jobs.Msg{Id: "src-1"}))

	// Assert
	require.NoError(t, err)
	assert.Equal(t, []string{string(jobs.SubjectFailed)}, pub.subjects)
}
