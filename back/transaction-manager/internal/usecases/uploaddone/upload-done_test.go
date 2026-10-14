package uploaddone_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/ismelen/inkomi/back/transaction-manager/internal/domain/models"
	"github.com/ismelen/inkomi/back/transaction-manager/internal/usecases/mocks"
	"github.com/ismelen/inkomi/back/transaction-manager/internal/usecases/uploaddone"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func fileSource(filename string) models.Source {
	size := int64(100)
	return models.Source{
		ID:       "source-1",
		UserID:   1,
		Title:    "Title",
		Filename: filename,
		Size:     &size,
		IsFile:   true,
	}
}

func repoReturning(detail *models.SourceDetail) *mocks.MockSourceRepository {
	return &mocks.MockSourceRepository{
		GetDetailByIDAndUserIDFn: func(ctx context.Context, id string, userID int) (*models.SourceDetail, error) {
			return detail, nil
		},
	}
}

func cloudWithFile(found bool) *mocks.MockCloudStorage {
	return &mocks.MockCloudStorage{
		CheckFn: func(id, ext string) bool { return found },
	}
}

// publish runs the use case and returns the published subject and decoded payload.
func publish(t *testing.T, detail *models.SourceDetail) (string, map[string]any) {
	t.Helper()

	var subject string
	var payload []byte
	queue := &mocks.MockQueue{
		PublishFn: func(ctx context.Context, s string, p []byte) error {
			subject, payload = s, p
			return nil
		},
	}
	uc := uploaddone.NewUploadDoneUC(repoReturning(detail), cloudWithFile(true), queue)

	require.NoError(t, uc.Execute(context.Background(), 1, "source-1"))

	var msg map[string]any
	require.NoError(t, json.Unmarshal(payload, &msg))
	return subject, msg
}

func TestUploadDoneUC_Execute_SourceNotFound_ReturnsError(t *testing.T) {
	t.Parallel()

	// Arrange
	uc := uploaddone.NewUploadDoneUC(&mocks.MockSourceRepository{}, cloudWithFile(true), &mocks.MockQueue{})

	// Act
	err := uc.Execute(context.Background(), 1, "source-1")

	// Assert
	assert.ErrorIs(t, err, models.ErrSourceNotFound)
}

func TestUploadDoneUC_Execute_PublishesCorrectSubject(t *testing.T) {
	t.Parallel()

	joinParent := &models.SourceParent{ID: "folder-1", Title: "Folder", Join: true}
	noJoinParent := &models.SourceParent{ID: "folder-1", Title: "Folder"}

	cases := []struct {
		name            string
		filename        string
		size            *int64
		kepubify        bool
		parent          *models.SourceParent
		expectedSubject string
	}{
		{"Library", "md5", nil, false, nil, models.SubjectJobLibrary},
		{"Manga", "chapter.cbz", new(int64(1)), false, nil, models.SubjectJobManga},
		{"Manga in join folder", "chapter.cbz", new(int64(1)), false, joinParent, models.SubjectJobManga},
		{"Epub in join folder", "book.epub", new(int64(1)), true, joinParent, models.SubjectJobWaitJoin},
		{"Epub in folder without join", "book.epub", new(int64(1)), true, noJoinParent, models.SubjectJobKepub},
		{"Epub kepubify", "book.epub", new(int64(1)), true, nil, models.SubjectJobKepub},
		{"Epub no kepubify", "book.epub", new(int64(1)), false, nil, models.SubjectJobDone},
		{"Already kepub", "book.kepub.epub", new(int64(1)), true, nil, models.SubjectJobDone},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			src := fileSource(tc.filename)
			src.Size = tc.size
			src.Kepubify = tc.kepubify
			detail := &models.SourceDetail{Source: src, Parent: tc.parent}

			// Act
			subject, _ := publish(t, detail)

			// Assert
			assert.Equal(t, tc.expectedSubject, subject)
		})
	}
}

func TestUploadDoneUC_Execute_Epub_PublishesBaseMessage(t *testing.T) {
	t.Parallel()

	// Arrange
	src := fileSource("book.epub")
	src.Kepubify = true
	src.ToCloud = true

	// Act
	_, msg := publish(t, &models.SourceDetail{Source: src})

	// Assert
	assert.Equal(t, map[string]any{
		"id":       "source-1",
		"userId":   float64(1),
		"filename": "book.epub",
		"title":    "Title",
		"kepubify": true,
		"toCloud":  true,
	}, msg)
}

func TestUploadDoneUC_Execute_Manga_AddsRtlAndMangaCfg(t *testing.T) {
	t.Parallel()

	// Arrange
	src := fileSource("chapter.cbz")
	src.IsRTL = true
	detail := &models.SourceDetail{
		Source: src,
		Config: &models.Config{Hash: "hash", Data: `{"gamma":1.5}`},
	}

	// Act
	_, msg := publish(t, detail)

	// Assert
	assert.Equal(t, true, msg["rtl"])
	assert.Equal(t, map[string]any{"gamma": 1.5}, msg["mangaCfg"])
	assert.NotContains(t, msg, "parent")
}

func TestUploadDoneUC_Execute_JoinFolder_AddsParent(t *testing.T) {
	t.Parallel()

	// Arrange
	detail := &models.SourceDetail{
		Source: fileSource("book.epub"),
		Parent: &models.SourceParent{ID: "folder-1", Title: "Folder", IsRTL: true, Join: true},
	}

	// Act
	_, msg := publish(t, detail)

	// Assert
	assert.Equal(t, map[string]any{
		"parentId": "folder-1",
		"title":    "Folder",
		"rtl":      true,
	}, msg["parent"])
	assert.NotContains(t, msg, "mangaCfg")
}

func TestUploadDoneUC_Execute_Success_SetsQueuedStatus(t *testing.T) {
	t.Parallel()

	// Arrange
	var status models.SourceStatus
	repo := repoReturning(&models.SourceDetail{Source: fileSource("book.epub")})
	repo.UpdateStatusFn = func(ctx context.Context, id string, userID int, s models.SourceStatus) error {
		status = s
		return nil
	}
	uc := uploaddone.NewUploadDoneUC(repo, cloudWithFile(true), &mocks.MockQueue{})

	// Act
	err := uc.Execute(context.Background(), 1, "source-1")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, models.SourceStatusQueued, status)
}

func TestUploadDoneUC_Execute_QueueError_ReturnsError(t *testing.T) {
	t.Parallel()

	// Arrange
	queueErr := errors.New("queue error")
	queue := &mocks.MockQueue{
		PublishFn: func(ctx context.Context, subject string, payload []byte) error {
			return queueErr
		},
	}
	repo := repoReturning(&models.SourceDetail{Source: fileSource("chapter.cbz")})
	uc := uploaddone.NewUploadDoneUC(repo, cloudWithFile(true), queue)

	// Act
	err := uc.Execute(context.Background(), 1, "source-1")

	// Assert
	assert.ErrorIs(t, err, queueErr)
}

func TestUploadDoneUC_Execute_SourceNotInCloud_ReturnsError(t *testing.T) {
	t.Parallel()

	// Arrange
	repo := repoReturning(&models.SourceDetail{Source: fileSource("chapter.cbz")})
	uc := uploaddone.NewUploadDoneUC(repo, cloudWithFile(false), &mocks.MockQueue{})

	// Act
	err := uc.Execute(context.Background(), 1, "source-1")

	// Assert
	assert.ErrorIs(t, err, models.ErrSourceNotUploaded)
}

func TestUploadDoneUC_Execute_Library_SkipsCloudCheck(t *testing.T) {
	t.Parallel()

	// Arrange
	src := fileSource("md5")
	src.Size = nil
	repo := repoReturning(&models.SourceDetail{Source: src})
	uc := uploaddone.NewUploadDoneUC(repo, cloudWithFile(false), &mocks.MockQueue{})

	// Act
	err := uc.Execute(context.Background(), 1, "source-1")

	// Assert
	assert.NoError(t, err)
}
