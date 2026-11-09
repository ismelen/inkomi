package retryupload_test

import (
	"context"
	"testing"

	"github.com/ismelen/inkomi/back/monolith/internal/modules/transactions/domain/models"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/transactions/usecases/mocks"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/transactions/usecases/retryupload"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func source(status models.SourceStatus, isFile bool, size int64) *models.Source {
	return &models.Source{Id: "src-1", UserId: 7, Filename: "c62.cbz", Status: status, IsFile: isFile, Size: &size}
}

func TestRetryUploadUC_Execute_PendingFile_ReturnsNewURL(t *testing.T) {
	t.Parallel()

	// Arrange
	repo := &mocks.MockSourceRepository{GetByIdAndUserIdFn: func(ctx context.Context, id string, userId int) (*models.Source, error) {
		return source(models.SourceStatusPendingUpload, true, 100), nil
	}}
	cloud := &mocks.MockCloudStorage{GetUrlFn: func(id, filename string) (string, error) {
		return "https://r2/" + id + "/" + filename, nil
	}}

	// Act
	url, err := retryupload.NewRetryUploadUC(repo, cloud).Execute(context.Background(), 7, "src-1")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "https://r2/src-1/c62.cbz", url)
}

func TestRetryUploadUC_Execute_NotRetryable_ReturnsError(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		src  *models.Source
	}{
		{"Already confirmed", source(models.SourceStatusQueued, true, 100)},
		{"Failed processing", source(models.SourceStatusFailed, true, 100)},
		{"Folder", source(models.SourceStatusPendingUpload, false, 0)},
		{"Library item", source(models.SourceStatusPendingUpload, true, 0)},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			repo := &mocks.MockSourceRepository{GetByIdAndUserIdFn: func(ctx context.Context, id string, userId int) (*models.Source, error) {
				return tc.src, nil
			}}

			// Act
			_, err := retryupload.NewRetryUploadUC(repo, &mocks.MockCloudStorage{}).Execute(context.Background(), 7, "src-1")

			// Assert
			assert.ErrorIs(t, err, models.ErrSourceNotRetryable)
		})
	}
}
