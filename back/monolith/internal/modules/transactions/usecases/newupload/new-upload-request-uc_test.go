package newupload_test

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/ismelen/inkomi/back/monolith/internal/modules/transactions/domain/models"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/transactions/infra/api/dtos"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/transactions/usecases/mocks"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/transactions/usecases/newupload"
	"github.com/stretchr/testify/assert"
)

func TestNewUploadRequestUC_Execute_InvalidEreaderKey_ReturnsError(t *testing.T) {
	// Arrange
	uc := newupload.NewNewUploadRequestUC(
		&mocks.MockTxManager{},
		&mocks.MockSourceRepository{},
		&mocks.MockConfigRepository{},
		&mocks.MockCloudStorage{},
		&mocks.MockStorageQuota{},
		models.StorageEstimateOptions{},
	)

	req := dtos.UploadRequestDTO{
		EReaderKey: "INVALID_KEY",
	}

	// Act
	_, err := uc.Execute(context.Background(), req, 1)

	// Assert
	assert.Error(t, err, "expected error for invalid ereader key")
}

func TestNewUploadRequestUC_Execute_ValidRequest_ReturnsSources(t *testing.T) {
	// Arrange
	txManager := &mocks.MockTxManager{}
	sourceRepo := &mocks.MockSourceRepository{}
	configRepo := &mocks.MockConfigRepository{}
	cloudStorage := &mocks.MockCloudStorage{
		GetUrlFn: func(id string, filename string, size int64) (string, error) {
			return "http://example.com/" + id + "/" + filename, nil
		},
	}

	uc := newupload.NewNewUploadRequestUC(
		txManager,
		sourceRepo,
		configRepo,
		cloudStorage,
		&mocks.MockStorageQuota{},
		models.StorageEstimateOptions{},
	)

	req := dtos.UploadRequestDTO{
		EReaderKey: "K1",
		Sources: []dtos.SourceDTO{
			{
				NoItemsSourceDTO: dtos.NoItemsSourceDTO{
					Title:    "Source 1",
					Filename: "test1.zip",
					Type:     models.SourceTypeFile,
				},
			},
		},
	}

	// Act
	sources, err := uc.Execute(context.Background(), req, 1)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, sources, 1)
	assert.Equal(t, "http://example.com//test1.zip", sources[0].Url) // Id is empty because mock returns empty
}

func TestNewUploadRequestUC_Execute_WithConfigs_SavesConfigsAndSources(t *testing.T) {
	// Arrange
	txManager := &mocks.MockTxManager{}
	configHashCreated := ""
	configRepo := &mocks.MockConfigRepository{
		CreateFn: func(ctx context.Context, config *models.Config) (string, error) {
			configHashCreated = config.Hash
			return config.Hash, nil
		},
	}
	sourceRepo := &mocks.MockSourceRepository{}
	cloudStorage := &mocks.MockCloudStorage{}

	uc := newupload.NewNewUploadRequestUC(
		txManager,
		sourceRepo,
		configRepo,
		cloudStorage,
		&mocks.MockStorageQuota{},
		models.StorageEstimateOptions{},
	)

	confKey := "conf1"
	req := dtos.UploadRequestDTO{
		EReaderKey: "K1",
		Configs: map[string]dtos.ConfigDTO{
			confKey: {
				EReaderKey: "K1",
			},
		},
		Sources: []dtos.SourceDTO{
			{
				NoItemsSourceDTO: dtos.NoItemsSourceDTO{
					Title:     "Source Config",
					Filename:  "test2.zip",
					Type:      models.SourceTypeFile,
					ConfigKey: &confKey,
				},
			},
		},
	}

	// Act
	_, err := uc.Execute(context.Background(), req, 1)

	// Assert
	assert.NoError(t, err)
	assert.NotEmpty(t, configHashCreated, "expected config to be saved and hash generated")
}

func TestNewUploadRequestUC_Execute_SourceWithItemsNotFolder_SetsFolderAsType(t *testing.T) {
	// Arrange
	uc := newupload.NewNewUploadRequestUC(
		&mocks.MockTxManager{},
		&mocks.MockSourceRepository{},
		&mocks.MockConfigRepository{},
		&mocks.MockCloudStorage{},
		&mocks.MockStorageQuota{},
		models.StorageEstimateOptions{},
	)

	req := dtos.UploadRequestDTO{
		EReaderKey: "K1",
		Sources: []dtos.SourceDTO{
			{
				NoItemsSourceDTO: dtos.NoItemsSourceDTO{
					Type: models.SourceTypeFile,
				},
				Items: []dtos.NoItemsSourceDTO{
					{Type: models.SourceTypeFile},
				},
			},
		},
	}

	// Act
	srcs, err := uc.Execute(context.Background(), req, 1)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, srcs[0].Type, models.SourceTypeFolder)
}

func TestNewUploadRequestUC_Execute_ChildInheritsConfigHash_UsesParentHash(t *testing.T) {
	// Arrange
	txManager := &mocks.MockTxManager{}
	configRepo := &mocks.MockConfigRepository{
		CreateFn: func(ctx context.Context, config *models.Config) (string, error) {
			return config.Hash, nil
		},
	}

	var capturedChild *models.Source
	sourceRepo := &mocks.MockSourceRepository{
		CreateFn: func(ctx context.Context, source *models.Source) (string, error) {
			if source.ParentId != nil {
				capturedChild = source
			}
			return "some-id", nil
		},
	}
	cloudStorage := &mocks.MockCloudStorage{}

	uc := newupload.NewNewUploadRequestUC(
		txManager,
		sourceRepo,
		configRepo,
		cloudStorage,
		&mocks.MockStorageQuota{},
		models.StorageEstimateOptions{},
	)

	parentConfKey := "parent-conf"
	req := dtos.UploadRequestDTO{
		EReaderKey: "K1",
		Configs: map[string]dtos.ConfigDTO{
			parentConfKey: {EReaderKey: "K1"},
		},
		Sources: []dtos.SourceDTO{
			{
				NoItemsSourceDTO: dtos.NoItemsSourceDTO{
					Type:      models.SourceTypeFolder,
					ConfigKey: &parentConfKey,
				},
				Items: []dtos.NoItemsSourceDTO{
					{
						Type:     models.SourceTypeFile,
						Filename: "child.zip",
					},
				},
			},
		},
	}

	// Act
	_, err := uc.Execute(context.Background(), req, 1)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, capturedChild)
	assert.NotNil(t, capturedChild.ConfigKey)
}

func TestNewUploadRequestUC_Execute_ChildInheritsParentAttributes_MatchesParent(t *testing.T) {
	// Arrange
	txManager := &mocks.MockTxManager{}
	configRepo := &mocks.MockConfigRepository{}

	var capturedChild *models.Source
	sourceRepo := &mocks.MockSourceRepository{
		CreateFn: func(ctx context.Context, source *models.Source) (string, error) {
			if source.ParentId != nil {
				capturedChild = source
			}
			return "parent-id", nil
		},
	}
	cloudStorage := &mocks.MockCloudStorage{}

	uc := newupload.NewNewUploadRequestUC(
		txManager,
		sourceRepo,
		configRepo,
		cloudStorage,
		&mocks.MockStorageQuota{},
		models.StorageEstimateOptions{},
	)

	req := dtos.UploadRequestDTO{
		EReaderKey: "KoMT",
		Sources: []dtos.SourceDTO{
			{
				ShouldJoin: true,
				NoItemsSourceDTO: dtos.NoItemsSourceDTO{
					Type:             models.SourceTypeFolder,
					ReadingDirection: models.ReadingDirectionRTL,
				},
				Items: []dtos.NoItemsSourceDTO{
					{
						Type: models.SourceTypeFile,
					},
				},
			},
		},
	}

	// Act
	_, err := uc.Execute(context.Background(), req, 1)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, capturedChild)
	assert.Equal(t, "parent-id", *capturedChild.ParentId)
	assert.True(t, capturedChild.Join)
	assert.True(t, capturedChild.IsRTL)
	assert.True(t, capturedChild.Kepubify)
}

func TestNewUploadRequestUC_Execute_FolderAndLibrary_EmptyUrl(t *testing.T) {
	// Arrange
	uc := newupload.NewNewUploadRequestUC(
		&mocks.MockTxManager{},
		&mocks.MockSourceRepository{},
		&mocks.MockConfigRepository{},
		&mocks.MockCloudStorage{
			GetUrlFn: func(id string, filename string, size int64) (string, error) {
				return "should-not-be-called", nil
			},
		},
		&mocks.MockStorageQuota{},
		models.StorageEstimateOptions{},
	)

	req := dtos.UploadRequestDTO{
		EReaderKey: "K1",
		Sources: []dtos.SourceDTO{
			{
				NoItemsSourceDTO: dtos.NoItemsSourceDTO{
					Type: models.SourceTypeFolder,
				},
				Items: []dtos.NoItemsSourceDTO{
					{
						Type: models.SourceTypeLibrary,
					},
				},
			},
		},
	}

	// Act
	sources, err := uc.Execute(context.Background(), req, 1)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, sources, 1)
	assert.Empty(t, sources[0].Url)
	assert.Len(t, sources[0].Items, 1)
	assert.Empty(t, sources[0].Items[0].Url)
}

func TestNewUploadRequestUC_Execute_JoinedFolder_ReservesItsPeak(t *testing.T) {
	// Arrange
	next := 0
	sourceRepo := &mocks.MockSourceRepository{
		CreateFn: func(ctx context.Context, source *models.Source) (string, error) {
			next++
			return fmt.Sprint("id-", next), nil
		},
	}
	var reserved []models.StorageReservation
	quota := &mocks.MockStorageQuota{ReserveFn: func(ctx context.Context, r []models.StorageReservation) error {
		reserved = r
		return nil
	}}
	opts := models.StorageEstimateOptions{EpubRatio: 1, Concurrency: 1}
	uc := newupload.NewNewUploadRequestUC(&mocks.MockTxManager{}, sourceRepo, &mocks.MockConfigRepository{}, &mocks.MockCloudStorage{}, quota, opts)
	size := int64(100)
	req := dtos.UploadRequestDTO{
		EReaderKey: "K1",
		Sources: []dtos.SourceDTO{{
			NoItemsSourceDTO: dtos.NoItemsSourceDTO{Title: "Folder", Filename: "Folder", Type: models.SourceTypeFolder},
			ShouldJoin:       true,
			Items: []dtos.NoItemsSourceDTO{
				{Title: "1", Filename: "1.cbz", Type: models.SourceTypeFile, Size: &size},
				{Title: "2", Filename: "2.cbz", Type: models.SourceTypeFile, Size: &size},
			},
		}},
	}

	// Act
	_, err := uc.Execute(context.Background(), req, 7)

	// Assert: join peak = 200 of epubs + a 200 block
	require.NoError(t, err)
	require.Len(t, reserved, 1)
	assert.Equal(t, models.StorageReservation{RootId: "id-1", UserId: 7, Bytes: 400}, reserved[0])
}

func TestNewUploadRequestUC_Execute_NoSpace_ReturnsInsufficientStorage(t *testing.T) {
	// Arrange
	quota := &mocks.MockStorageQuota{ReserveFn: func(ctx context.Context, r []models.StorageReservation) error {
		return models.ErrInsufficientStorage
	}}
	uc := newupload.NewNewUploadRequestUC(&mocks.MockTxManager{}, &mocks.MockSourceRepository{}, &mocks.MockConfigRepository{}, &mocks.MockCloudStorage{}, quota, models.StorageEstimateOptions{})
	size := int64(100)
	req := dtos.UploadRequestDTO{
		EReaderKey: "K1",
		Sources:    []dtos.SourceDTO{{NoItemsSourceDTO: dtos.NoItemsSourceDTO{Title: "a", Filename: "a.cbz", Type: models.SourceTypeFile, Size: &size}}},
	}

	// Act
	_, err := uc.Execute(context.Background(), req, 7)

	// Assert
	assert.ErrorIs(t, err, models.ErrInsufficientStorage)
}
