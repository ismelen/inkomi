package uploaddone_test

import (
	"context"
	"errors"
	"testing"

	"github.com/ismelen/inkomi/back/transaction-manager/internal/domain/models"
	"github.com/ismelen/inkomi/back/transaction-manager/internal/usecases/mocks"
	"github.com/ismelen/inkomi/back/transaction-manager/internal/usecases/uploaddone"
	"github.com/stretchr/testify/assert"
)

func TestUploadDoneUC_Execute_SourceNotFound_ReturnsError(t *testing.T) {
	// Arrange
	mockRepo := &mocks.MockSourceRepository{
		GetByIdAndUserIdCompactFn: func(ctx context.Context, id string, userID int) (*models.CompactSoruce, error) {
			return nil, errors.New("not found")
		},
	}
	mockQueue := &mocks.MockQueue{}
	mockCloudStorage := &mocks.MockCloudStorage{
		CheckFn: func(id, ext string) bool { return true },
	}
	uc := uploaddone.NewUploadDoneUC(mockRepo, mockCloudStorage, mockQueue)

	// Act
	err := uc.Execute(context.Background(), 1, "source-1")

	// Assert
	assert.EqualError(t, err, "not found")
}

func TestUploadDoneUC_Execute_PublishesCorrectSubject(t *testing.T) {
	cases := []struct {
		name            string
		sourceType      models.SourceType
		filename        string
		kepubify        bool
		expectedSubject string
	}{
		{"Library", models.SourceTypeLibrary, "", false, "job.init.library"},
		{"Kepub", models.SourceTypeFile, "book.kepub.epub", false, "job.step.done"},
		{"Epub Kepubify True", models.SourceTypeFile, "book.epub", true, "job.step.kepub"},
		{"Epub Kepubify False", models.SourceTypeFile, "book.epub", false, "job.step.done"},
		{"Manga File", models.SourceTypeFile, "chapter.zip", false, "job.init.manga"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			mockRepo := &mocks.MockSourceRepository{
				GetByIdAndUserIdCompactFn: func(ctx context.Context, id string, userID int) (*models.CompactSoruce, error) {
					return &models.CompactSoruce{
						Id:       "source-1",
						Type:     tc.sourceType,
						Filename: tc.filename,
						Kepubify: tc.kepubify,
					}, nil
				},
				UpdateStatusFn: func(ctx context.Context, id string, userId int, status models.SourceStatus, errMsg *string) error {
					assert.Equal(t, models.SourceStatusQueued, status)
					return nil
				},
			}

			publishedSubject := ""
			mockQueue := &mocks.MockQueue{
				PublishFn: func(ctx context.Context, subject string, payload []byte) error {
					publishedSubject = subject
					return nil
				},
			}

			mockCloudStorage := &mocks.MockCloudStorage{
				CheckFn: func(id, ext string) bool { return true },
			}

			uc := uploaddone.NewUploadDoneUC(mockRepo, mockCloudStorage, mockQueue)

			// Act
			err := uc.Execute(context.Background(), 1, "source-1")

			// Assert
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedSubject, publishedSubject)
		})
	}
}

func TestUploadDoneUC_Execute_QueueError_ReturnsError(t *testing.T) {
	// Arrange
	mockRepo := &mocks.MockSourceRepository{
		GetByIdAndUserIdCompactFn: func(ctx context.Context, id string, userID int) (*models.CompactSoruce, error) {
			return &models.CompactSoruce{Id: "source-1", Type: models.SourceTypeFile, Filename: "chapter.zip"}, nil
		},
	}

	mockQueue := &mocks.MockQueue{
		PublishFn: func(ctx context.Context, subject string, payload []byte) error {
			return errors.New("queue error")
		},
	}

	mockCloudStorage := &mocks.MockCloudStorage{
		CheckFn: func(id, ext string) bool { return true },
	}
	uc := uploaddone.NewUploadDoneUC(mockRepo, mockCloudStorage, mockQueue)

	// Act
	err := uc.Execute(context.Background(), 1, "source-1")

	// Assert
	assert.EqualError(t, err, "queue error")
}

func TestUploadDoneUC_Execute_SourceNotInCloud_ReturnsError(t *testing.T) {
	// Arrange
	mockRepo := &mocks.MockSourceRepository{
		GetByIdAndUserIdCompactFn: func(ctx context.Context, id string, userID int) (*models.CompactSoruce, error) {
			return &models.CompactSoruce{Id: "source-1", Type: models.SourceTypeFile, Filename: "chapter.zip"}, nil
		},
		UpdateStatusFn: func(ctx context.Context, id string, userId int, status models.SourceStatus, errMsg *string) error {
			return nil
		},
	}

	mockQueue := &mocks.MockQueue{
		PublishFn: func(ctx context.Context, subject string, payload []byte) error {
			return nil
		},
	}

	mockCloudStorage := &mocks.MockCloudStorage{
		CheckFn: func(id, ext string) bool { return false }, // Simulates source not found in cloud
	}

	uc := uploaddone.NewUploadDoneUC(mockRepo, mockCloudStorage, mockQueue)

	// Act
	err := uc.Execute(context.Background(), 1, "source-1")

	// Assert
	assert.EqualError(t, err, "source doesn't exists")
}
