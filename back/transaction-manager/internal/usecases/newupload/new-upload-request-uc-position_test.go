package newupload_test

import (
	"context"
	"testing"

	"github.com/ismelen/inkomi/back/transaction-manager/internal/domain/models"
	"github.com/ismelen/inkomi/back/transaction-manager/internal/infra/api/dtos"
	"github.com/ismelen/inkomi/back/transaction-manager/internal/usecases/mocks"
	"github.com/ismelen/inkomi/back/transaction-manager/internal/usecases/newupload"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewUploadRequestUC_Execute_FolderItems_SavesPositionInOrder(t *testing.T) {
	t.Parallel()

	// Arrange
	positions := map[string]int{}
	sourceRepo := &mocks.MockSourceRepository{
		CreateFn: func(ctx context.Context, source *models.Source) (string, error) {
			positions[source.Title] = source.Position
			return source.Title, nil
		},
	}
	uc := newupload.NewNewUploadRequestUC(&mocks.MockTxManager{}, sourceRepo, &mocks.MockConfigRepository{}, &mocks.MockCloudStorage{})

	req := dtos.UploadRequestDTO{
		EReaderKey: "KoMT",
		Sources: []dtos.SourceDTO{
			{
				NoItemsSourceDTO: dtos.NoItemsSourceDTO{Title: "Folder", Type: models.SourceTypeFolder},
				Items: []dtos.NoItemsSourceDTO{
					{Title: "Ch 1", Type: models.SourceTypeFile},
					{Title: "Ch 2", Type: models.SourceTypeFile},
					{Title: "Ch 3", Type: models.SourceTypeFile},
				},
			},
		},
	}

	// Act
	_, err := uc.Execute(context.Background(), req, 1)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, map[string]int{"Folder": 0, "Ch 1": 0, "Ch 2": 1, "Ch 3": 2}, positions)
}
