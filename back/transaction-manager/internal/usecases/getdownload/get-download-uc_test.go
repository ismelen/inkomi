package getdownload_test

import (
	"context"
	"testing"
	"time"

	"github.com/ismelen/inkomi/back/transaction-manager/internal/domain/models"
	"github.com/ismelen/inkomi/back/transaction-manager/internal/usecases/getdownload"
	"github.com/ismelen/inkomi/back/transaction-manager/internal/usecases/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func repoWith(source *models.Source) *mocks.MockSourceRepository {
	return &mocks.MockSourceRepository{
		GetByIdAndUserIdFn: func(ctx context.Context, id string, userId int) (*models.Source, error) {
			return source, nil
		},
	}
}

func TestGetDownloadUC_Execute_NotDownloadable_ReturnsError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		repo *mocks.MockSourceRepository
		want error
	}{
		{"not found", &mocks.MockSourceRepository{}, models.ErrSourceNotFound},
		{"processing", repoWith(&models.Source{Id: "s", IsFile: true, Status: models.SourceStatusProcessing}), models.ErrSourceNotReady},
		{"failed", repoWith(&models.Source{Id: "s", IsFile: true, Status: models.SourceStatusFailed}), models.ErrSourceNotReady},
		{"joined child", repoWith(&models.Source{Id: "s", IsFile: true, Status: models.SourceStatusJoined}), models.ErrSourceNotReady},
		{"folder", repoWith(&models.Source{Id: "s", IsFile: false, Status: models.SourceStatusDone}), models.ErrSourceNotReady},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			uc := getdownload.NewGetDownloadUC(tt.repo, &mocks.MockCloudStorage{CheckFn: func(id, filename string) bool { return true }})

			// Act
			download, err := uc.Execute(context.Background(), 1, "s")

			// Assert
			assert.ErrorIs(t, err, tt.want)
			assert.Nil(t, download)
		})
	}
}

func TestGetDownloadUC_Execute_MissingInStorage_ReturnsFileNotAvailable(t *testing.T) {
	t.Parallel()

	// Arrange
	repo := repoWith(&models.Source{Id: "s", Filename: "book.epub", IsFile: true, Status: models.SourceStatusSent})
	uc := getdownload.NewGetDownloadUC(repo, &mocks.MockCloudStorage{})

	// Act
	download, err := uc.Execute(context.Background(), 1, "s")

	// Assert
	assert.ErrorIs(t, err, models.ErrFileNotAvailable)
	assert.Nil(t, download)
}

func TestGetDownloadUC_Execute_Sent_ReturnsPresignedUrl(t *testing.T) {
	t.Parallel()

	// Arrange
	size := int64(1024)
	expiresAt := time.Date(2026, 10, 16, 10, 15, 0, 0, time.UTC)
	repo := repoWith(&models.Source{Id: "s", Filename: "book.epub", Size: &size, IsFile: true, IsResult: true, Status: models.SourceStatusSent})
	var gotKey string
	cloud := &mocks.MockCloudStorage{
		CheckFn: func(id, filename string) bool { return true },
		GetDownloadUrlFn: func(id, filename string) (string, time.Time, error) {
			gotKey = id + "/" + filename
			return "https://r2/s/book.epub", expiresAt, nil
		},
	}
	uc := getdownload.NewGetDownloadUC(repo, cloud)

	// Act
	download, err := uc.Execute(context.Background(), 1, "s")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "s/book.epub", gotKey)
	assert.Equal(t, "https://r2/s/book.epub", download.Url)
	assert.Equal(t, "book.epub", download.Filename)
	assert.Equal(t, &size, download.Size)
	assert.Equal(t, expiresAt, download.ExpiresAt)
}
